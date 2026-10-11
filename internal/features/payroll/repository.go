package payroll

import (
	"context"

	"gorm.io/gorm"
)

type ComponentFilter struct {
	Page    int64
	PerPage int64
	Q       string
	Type    string
	Active  *bool
}

type PeriodFilter struct {
	Page   int64
	PerPage int64
	Q      string
	Status string
	Year   int
}

// EmployeeComponentRow memuat hasil join employee + pay_components.
type EmployeeComponentRow struct {
	EmployeeComponent
	EmployeeName  string `gorm:"column:employee_name"`
	EmployeeNo    string `gorm:"column:employee_no"`
	ComponentCode string `gorm:"column:component_code"`
	ComponentName string `gorm:"column:component_name"`
	ComponentType string `gorm:"column:component_type"`
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func (r *Repository) ListComponents(ctx context.Context, f ComponentFilter) ([]*PayComponent, int64, error) {
	query := r.dbCtx(ctx).Model(&PayComponent{})
	if f.Q != "" {
		like := "%" + f.Q + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	if f.Type != "" {
		query = query.Where("type = ?", f.Type)
	}
	if f.Active != nil {
		query = query.Where("is_active = ?", *f.Active)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*PayComponent
	err := query.Order("type ASC, code ASC").
		Offset(int((f.Page - 1) * f.PerPage)).
		Limit(int(f.PerPage)).
		Find(&list).Error
	return list, total, err
}

func (r *Repository) FindComponent(ctx context.Context, id int64) (*PayComponent, error) {
	var c PayComponent
	if err := r.dbCtx(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) FindComponentByCode(ctx context.Context, code string) (*PayComponent, error) {
	var c PayComponent
	if err := r.dbCtx(ctx).Where("code = ?", code).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) CreateComponent(ctx context.Context, c *PayComponent) error {
	return r.dbCtx(ctx).Create(c).Error
}

func (r *Repository) UpdateComponent(ctx context.Context, c *PayComponent) error {
	return r.dbCtx(ctx).Model(c).
		Select("code", "name", "type", "description", "is_active").
		Updates(c).Error
}

func (r *Repository) SoftDeleteComponent(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&PayComponent{}, id).Error
}

func (r *Repository) ComponentCodeExists(ctx context.Context, code string, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Unscoped().Model(&PayComponent{}).
		Where("code = ? AND id <> ?", code, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountComponentAssignments(ctx context.Context, componentID int64) (int64, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&EmployeeComponent{}).
		Where("component_id = ?", componentID).Count(&count).Error
	return count, err
}

const employeeComponentSelect = `
SELECT ec.id, ec.employee_id, ec.component_id, ec.calc_mode, ec.amount, ec.note, ec.created_at, ec.updated_at,
       e.full_name AS employee_name, e.employee_no AS employee_no,
       pc.code AS component_code, pc.name AS component_name, pc.type AS component_type
FROM employee_components ec
JOIN employees e ON e.id = ec.employee_id
JOIN pay_components pc ON pc.id = ec.component_id `

func (r *Repository) ListEmployeeComponents(ctx context.Context, employeeID int64, page, per int64) ([]*EmployeeComponentRow, int64, error) {
	where := ""
	args := []any{}
	if employeeID > 0 {
		where = "WHERE ec.employee_id = ?"
		args = append(args, employeeID)
	}

	var total int64
	if err := r.dbCtx(ctx).Raw(
		"SELECT COUNT(*) FROM employee_components ec "+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	list := []*EmployeeComponentRow{}
	err := r.dbCtx(ctx).Raw(
		employeeComponentSelect+where+` ORDER BY pc.type ASC, pc.code ASC LIMIT ? OFFSET ?`,
		append(args, per, (page-1)*per)...,
	).Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *Repository) FindEmployeeComponent(ctx context.Context, id int64) (*EmployeeComponentRow, error) {
	var row EmployeeComponentRow
	err := r.dbCtx(ctx).Raw(employeeComponentSelect+"WHERE ec.id = ?", id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *Repository) CreateEmployeeComponent(ctx context.Context, ec *EmployeeComponent) error {
	return r.dbCtx(ctx).Create(ec).Error
}

func (r *Repository) UpdateEmployeeComponent(ctx context.Context, ec *EmployeeComponent) error {
	return r.dbCtx(ctx).Model(ec).Select("calc_mode", "amount", "note").Updates(ec).Error
}

func (r *Repository) DeleteEmployeeComponent(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&EmployeeComponent{}, id).Error
}

func (r *Repository) EmployeeComponentExists(ctx context.Context, employeeID, componentID int64, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&EmployeeComponent{}).
		Where("employee_id = ? AND component_id = ? AND id <> ?", employeeID, componentID, excludeID).
		Count(&count).Error
	return count > 0, err
}

const periodAggregateSelect = `
SELECT p.id, p.name, p.period_month, p.period_year, p.status, p.paid_at, p.note, p.created_at, p.effective_weeks,
       COUNT(s.id) AS slip_count,
       COALESCE(SUM(s.total_earning), 0) AS total_earning,
       COALESCE(SUM(s.total_deduction), 0) AS total_deduction,
       COALESCE(SUM(s.take_home_pay), 0) AS total_take_home
FROM payroll_periods p
LEFT JOIN payslips s ON s.period_id = p.id `

type periodRow struct {
	PayrollPeriod
	SlipCount      int64 `gorm:"column:slip_count"`
	TotalEarning   int64 `gorm:"column:total_earning"`
	TotalDeduction int64 `gorm:"column:total_deduction"`
	TotalTakeHome  int64 `gorm:"column:total_take_home"`
}

func (r *Repository) listPeriodRows(ctx context.Context, where string, args []any, limit, offset int64) ([]*periodRow, error) {
	rows := []*periodRow{}
	err := r.dbCtx(ctx).Raw(
		periodAggregateSelect+where+` GROUP BY p.id ORDER BY p.period_year DESC, p.period_month DESC, p.id DESC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, err
}

func (r *Repository) countPeriods(ctx context.Context, where string, args []any) (int64, error) {
	var total int64
	err := r.dbCtx(ctx).Raw(
		"SELECT COUNT(*) FROM payroll_periods p "+where, args...,
	).Scan(&total).Error
	return total, err
}

func periodFilterWhere(f PeriodFilter) (string, []any) {
	where := "WHERE p.deleted_at IS NULL"
	args := []any{}
	if f.Status != "" {
		where += " AND p.status = ?"
		args = append(args, f.Status)
	}
	if f.Year > 0 {
		where += " AND p.period_year = ?"
		args = append(args, f.Year)
	}
	if f.Q != "" {
		where += " AND (p.name ILIKE ? OR CAST(p.period_year AS TEXT) ILIKE ?)"
		like := "%" + f.Q + "%"
		args = append(args, like, like)
	}
	return where, args
}

func (r *Repository) ListPeriods(ctx context.Context, f PeriodFilter) ([]*periodRow, int64, error) {
	where, args := periodFilterWhere(f)
	total, err := r.countPeriods(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.listPeriodRows(ctx, where, args, f.PerPage, (f.Page-1)*f.PerPage)
	return rows, total, err
}

func (r *Repository) FindPeriodRow(ctx context.Context, id int64) (*periodRow, error) {
	rows, err := r.listPeriodRows(ctx, "WHERE p.deleted_at IS NULL AND p.id = ?", []any{id}, 1, 0)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return rows[0], nil
}

func (r *Repository) FindPeriod(ctx context.Context, id int64) (*PayrollPeriod, error) {
	var p PayrollPeriod
	if err := r.dbCtx(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) SchoolWorkingDays(ctx context.Context) (int, error) {
	var wd int
	err := r.dbCtx(ctx).Raw(`SELECT working_days FROM school_profiles WHERE id = 1`).Scan(&wd).Error
	if err != nil || wd != 5 && wd != 6 {
		return 5, err
	}
	return wd, nil
}

func (r *Repository) CreatePeriod(ctx context.Context, p *PayrollPeriod) error {
	return r.dbCtx(ctx).Create(p).Error
}

func (r *Repository) UpdatePeriod(ctx context.Context, p *PayrollPeriod) error {
	return r.dbCtx(ctx).Model(p).Select("name", "note", "effective_weeks").Updates(p).Error
}

func (r *Repository) UpdatePeriodStatus(ctx context.Context, id int64, status string, setPaidAt bool) error {
	if setPaidAt {
		return r.dbCtx(ctx).Exec(
			"UPDATE payroll_periods SET status = ?, paid_at = now() WHERE id = ?", status, id,
		).Error
	}
	return r.dbCtx(ctx).Exec("UPDATE payroll_periods SET status = ? WHERE id = ?", status, id).Error
}

func (r *Repository) PeriodYearMonthExists(ctx context.Context, year, month int, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Unscoped().Model(&PayrollPeriod{}).
		Where("period_year = ? AND period_month = ? AND id <> ?", year, month, excludeID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) DeleteDraftPeriod(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`DELETE FROM payslip_items WHERE payslip_id IN (SELECT id FROM payslips WHERE period_id = ?)`, id,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM payslips WHERE period_id = ?`, id).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&PayrollPeriod{}, id).Error
	})
}
