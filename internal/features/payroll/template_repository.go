package payroll

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type templateRow struct {
	ID             int64     `gorm:"column:id"`
	Name           string    `gorm:"column:name"`
	Description    *string   `gorm:"column:description"`
	EmployeeType   *string   `gorm:"column:employee_type"`
	IsDefault      bool      `gorm:"column:is_default"`
	IsActive       bool      `gorm:"column:is_active"`
	ComponentCount int64     `gorm:"column:component_count"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

type templateComponentRow struct {
	ID            int64     `gorm:"column:id"`
	TemplateID    int64     `gorm:"column:template_id"`
	ComponentID   int64     `gorm:"column:component_id"`
	ComponentCode string    `gorm:"column:component_code"`
	ComponentName string    `gorm:"column:component_name"`
	ComponentType string    `gorm:"column:component_type"`
	CalcMode      string    `gorm:"column:calc_mode"`
	Amount        int64     `gorm:"column:amount"`
	Note          *string   `gorm:"column:note"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

const templateSelect = `
SELECT t.id, t.name, t.description, t.employee_type, t.is_default, t.is_active, t.created_at, t.updated_at,
       COALESCE(tc.cnt, 0) AS component_count
FROM payroll_templates t
LEFT JOIN (SELECT template_id, COUNT(*) AS cnt FROM payroll_template_components GROUP BY template_id) tc
       ON tc.template_id = t.id `

const templateComponentSelect = `
SELECT ptc.id, ptc.template_id, ptc.component_id, ptc.calc_mode, ptc.amount, ptc.note, ptc.created_at, ptc.updated_at,
       pc.code AS component_code, pc.name AS component_name, pc.type AS component_type
FROM payroll_template_components ptc
JOIN pay_components pc ON pc.id = ptc.component_id `

func (r *Repository) ListTemplates(ctx context.Context, q string, page, per int64) ([]*templateRow, int64, error) {
	where := "WHERE t.deleted_at IS NULL"
	args := []any{}
	if strings.TrimSpace(q) != "" {
		where += " AND (t.name ILIKE ? OR COALESCE(t.description,'') ILIKE ?)"
		p := "%" + q + "%"
		args = append(args, p, p)
	}

	var total int64
	if err := r.dbCtx(ctx).Raw(
		"SELECT COUNT(*) FROM payroll_templates t "+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	list := []*templateRow{}
	err := r.dbCtx(ctx).Raw(
		templateSelect+where+" ORDER BY t.is_default DESC, t.is_active DESC, t.name ASC LIMIT ? OFFSET ?",
		append(args, per, (page-1)*per)...,
	).Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *Repository) FindTemplate(ctx context.Context, id int64) (*templateRow, error) {
	var row templateRow
	err := r.dbCtx(ctx).Raw(templateSelect+"WHERE t.id = ? AND t.deleted_at IS NULL", id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *Repository) ListTemplateComponents(ctx context.Context, templateID int64) ([]*templateComponentRow, error) {
	list := []*templateComponentRow{}
	err := r.dbCtx(ctx).Raw(
		templateComponentSelect+"WHERE ptc.template_id = ? ORDER BY pc.type ASC, pc.code ASC", templateID,
	).Scan(&list).Error
	return list, err
}

func (r *Repository) CreateTemplate(ctx context.Context, t *PayrollTemplate) error {
	return r.dbCtx(ctx).Create(t).Error
}

func (r *Repository) CreateTemplateWithComponents(ctx context.Context, t *PayrollTemplate, comps []*PayrollTemplateComponent) error {
	return r.dbCtx(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		if len(comps) == 0 {
			return nil
		}
		for _, c := range comps {
			c.TemplateID = t.ID
		}
		return tx.Create(&comps).Error
	})
}

func (r *Repository) UpdateTemplateWithComponents(ctx context.Context, t *PayrollTemplate, comps []*PayrollTemplateComponent) error {
	return r.dbCtx(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(t).Select("name", "description", "employee_type", "is_default", "is_active").Updates(t).Error; err != nil {
			return err
		}
		if err := tx.Where("template_id = ?", t.ID).Delete(&PayrollTemplateComponent{}).Error; err != nil {
			return err
		}
		if len(comps) == 0 {
			return nil
		}
		for _, c := range comps {
			c.TemplateID = t.ID
		}
		return tx.Create(&comps).Error
	})
}

func (r *Repository) SoftDeleteTemplate(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&PayrollTemplate{}, id).Error
}

// ClearTemplateDefault menghapus tanda default dari template lain pada scope tipe
// pegawai yang sama (NULL employee_type dianggap scope "semua tipe").
func (r *Repository) ClearTemplateDefault(ctx context.Context, employeeType *string, excludeID int64) error {
	scope := "*"
	if employeeType != nil {
		scope = *employeeType
	}
	return r.dbCtx(ctx).Exec(
		`UPDATE payroll_templates SET is_default = false
		 WHERE id <> ? AND is_default AND deleted_at IS NULL AND COALESCE(employee_type, '*') = ?`,
		excludeID, scope,
	).Error
}

// ReplaceTemplateComponents mengganti seluruh isi komponen template.
func (r *Repository) ReplaceTemplateComponents(ctx context.Context, templateID int64, comps []*PayrollTemplateComponent) error {
	return r.dbCtx(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ?", templateID).Delete(&PayrollTemplateComponent{}).Error; err != nil {
			return err
		}
		if len(comps) == 0 {
			return nil
		}
		return tx.Create(&comps).Error
	})
}

// FindDefaultTemplate mencari template default aktif untuk tipe pegawai tsb;
// fallback ke template default dengan employee_type NULL (berlaku semua tipe).
func (r *Repository) FindDefaultTemplate(ctx context.Context, employeeType string) (*templateRow, error) {
	var row templateRow
	err := r.dbCtx(ctx).Raw(
		templateSelect+`WHERE t.deleted_at IS NULL AND t.is_default AND t.is_active
		               AND (t.employee_type = ? OR (t.employee_type IS NULL AND NOT EXISTS (
		                   SELECT 1 FROM payroll_templates d
		                   WHERE d.employee_type = ? AND d.is_default AND d.is_active AND d.deleted_at IS NULL
		               )))
		               ORDER BY t.employee_type NULLS LAST LIMIT 1`,
		employeeType, employeeType,
	).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *Repository) TemplateComponentExists(ctx context.Context, templateID, componentID, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&PayrollTemplateComponent{}).
		Where("template_id = ? AND component_id = ? AND id <> ?", templateID, componentID, excludeID).
		Count(&count).Error
	return count > 0, err
}

// ApplyTemplateToEmployee menyalin komponen template ke pegawai, melewati komponen
// yang sudah dimiliki pegawai (idempoten). Mengembalikan jumlah komponen yang diterapkan.
func (r *Repository) ApplyTemplateToEmployee(ctx context.Context, templateID, employeeID int64) (int64, error) {
	var result int64
	err := r.dbCtx(ctx).Transaction(func(tx *gorm.DB) error {
		comps := []*PayrollTemplateComponent{}
		if err := tx.Where("template_id = ?", templateID).Find(&comps).Error; err != nil {
			return err
		}
		if len(comps) == 0 {
			return nil
		}
		var existing []int64
		if err := tx.Model(&EmployeeComponent{}).
			Where("employee_id = ?", employeeID).
			Pluck("component_id", &existing).Error; err != nil {
			return err
		}
		has := map[int64]bool{}
		for _, id := range existing {
			has[id] = true
		}
		for _, c := range comps {
			if has[c.ComponentID] {
				continue
			}
			ec := &EmployeeComponent{
				EmployeeID: employeeID, ComponentID: c.ComponentID,
				CalcMode: c.CalcMode, Amount: c.Amount, Note: c.Note,
			}
			if err := tx.Create(ec).Error; err != nil {
				return err
			}
			result++
		}
		return nil
	})
	return result, err
}

// DeleteEmployeeComponents menghapus seluruh komponen pegawai (dipakai mode replace).
func (r *Repository) DeleteEmployeeComponents(ctx context.Context, employeeID int64) (int64, error) {
	res := r.dbCtx(ctx).Where("employee_id = ?", employeeID).Delete(&EmployeeComponent{})
	return res.RowsAffected, res.Error
}

// ListActiveEmployeeIDs mengembalikan id seluruh pegawai aktif, opsional difilter tipe.
func (r *Repository) ListActiveEmployeeIDs(ctx context.Context, employeeType string) ([]int64, error) {
	q := r.dbCtx(ctx).Table("employees").Where("is_active = ? AND deleted_at IS NULL", true)
	if employeeType != "" {
		q = q.Where("employee_type = ?", employeeType)
	}
	var ids []int64
	err := q.Pluck("id", &ids).Error
	return ids, err
}
