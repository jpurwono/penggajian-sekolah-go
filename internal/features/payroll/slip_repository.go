package payroll

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type employeeLite struct {
	ID         int64  `gorm:"column:id"`
	EmployeeNo string `gorm:"column:employee_no"`
	FullName   string `gorm:"column:full_name"`
	JobTitle   *string `gorm:"column:job_title"`
	BaseSalary int64  `gorm:"column:base_salary"`
}

type assignmentRow struct {
	EmployeeID    int64   `gorm:"column:employee_id"`
	ComponentID   int64   `gorm:"column:component_id"`
	Code          string  `gorm:"column:code"`
	Name          string  `gorm:"column:name"`
	Type          string  `gorm:"column:type"`
	CalcMode      string  `gorm:"column:calc_mode"`
	Amount        int64   `gorm:"column:amount"`
	Note          *string `gorm:"column:note"`
}

type SlipRepository struct {
	db *gorm.DB
}

func NewSlipRepository(db *gorm.DB) *SlipRepository {
	return &SlipRepository{db: db}
}

func (r *SlipRepository) ListActiveEmployees(ctx context.Context) ([]*employeeLite, error) {
	list := []*employeeLite{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, employee_no, full_name, job_title, base_salary
		FROM employees WHERE deleted_at IS NULL AND is_active = true
		ORDER BY full_name ASC
	`).Scan(&list).Error
	return list, err
}

// ListAssignments memuat semua komponen recurring pegawai aktif beserta data komponen untuk snapshot.
func (r *SlipRepository) ListAssignments(ctx context.Context) ([]*assignmentRow, error) {
	list := []*assignmentRow{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT ec.employee_id, ec.component_id, pc.code, pc.name, pc.type, ec.calc_mode, ec.amount, ec.note
		FROM employee_components ec
		JOIN employees e ON e.id = ec.employee_id AND e.deleted_at IS NULL AND e.is_active = true
		JOIN pay_components pc ON pc.id = ec.component_id
		ORDER BY e.full_name ASC, pc.type ASC, pc.code ASC
	`).Scan(&list).Error
	return list, err
}

// ListPeriodJam mempertahankan data jam saat slip digenerate ulang.
type slipJamRow struct {
	EmployeeID    int64 `gorm:"column:employee_id"`
	JamMengajar   *int  `gorm:"column:jam_mengajar"`
	JamTidakHadir *int  `gorm:"column:jam_tidak_hadir"`
	HariHadir     *int  `gorm:"column:hari_hadir"`
}

func (r *SlipRepository) ListPeriodJam(ctx context.Context, tx *gorm.DB, periodID int64) (map[int64]*slipJamRow, error) {
	rows := []*slipJamRow{}
	err := tx.WithContext(ctx).Raw(`
		SELECT employee_id, jam_mengajar, jam_tidak_hadir, hari_hadir
		FROM payslips WHERE period_id = ?
	`, periodID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[int64]*slipJamRow{}
	for _, row := range rows {
		out[row.EmployeeID] = row
	}
	return out, nil
}

type autoItemTarget struct {
	ID      int64  `gorm:"column:id"`
	PayslipID int64 `gorm:"column:payslip_id"`
	Amount  int64  `gorm:"column:amount"`
	Note    *string `gorm:"column:note"`
}

// FindAutoItem mencari item hasil generate (is_auto) milik komponen tertentu pada satu slip.
func (r *SlipRepository) FindAutoItem(ctx context.Context, tx *gorm.DB, payslipID, componentID int64) (*autoItemTarget, error) {
	row := &autoItemTarget{}
	err := tx.WithContext(ctx).Raw(`
		SELECT id, payslip_id, amount, note FROM payslip_items
		WHERE payslip_id = ? AND component_id = ? AND is_auto = true
		ORDER BY id ASC LIMIT 1
	`, payslipID, componentID).Scan(row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	return row, nil
}

func (r *SlipRepository) UpdateAutoItemAmount(ctx context.Context, tx *gorm.DB, itemID int64, amount int64, note *string) error {
	return tx.WithContext(ctx).Exec(`
		UPDATE payslip_items SET amount = ?, note = ? WHERE id = ?
	`, amount, note, itemID).Error
}

func (r *SlipRepository) CreateItemTx(ctx context.Context, tx *gorm.DB, item *PayslipItem) error {
	return tx.WithContext(ctx).Create(item).Error
}

// DeleteOrphanAutoItems menghapus item otomatis yang komponennya sudah tidak lagi
// melekat pada pegawai (kecuali GP yang bersumber dari gaji pokok).
func (r *SlipRepository) DeleteOrphanAutoItems(ctx context.Context, tx *gorm.DB, payslipID int64, keepComponentIDs []int64) (int64, error) {
	query := tx.WithContext(ctx).Where(
		"payslip_id = ? AND is_auto = true AND code_snapshot <> ?", payslipID, ComponentCodeGajiPokok,
	)
	if len(keepComponentIDs) > 0 {
		query = query.Where("component_id NOT IN ?", keepComponentIDs)
	}
	res := query.Delete(&PayslipItem{})
	return res.RowsAffected, res.Error
}

func (r *SlipRepository) DeletePeriodSlips(ctx context.Context, tx *gorm.DB, periodID int64) error {
	if err := tx.Exec(
		`DELETE FROM payslip_items WHERE payslip_id IN (SELECT id FROM payslips WHERE period_id = ?)`, periodID,
	).Error; err != nil {
		return err
	}
	return tx.Exec(`DELETE FROM payslips WHERE period_id = ?`, periodID).Error
}

func (r *SlipRepository) CreateSlipWithItems(tx *gorm.DB, slip *Payslip, items []*PayslipItem) error {
	if err := tx.Create(slip).Error; err != nil {
		return err
	}
	for _, item := range items {
		item.PayslipID = slip.ID
	}
	if len(items) > 0 {
		return tx.Create(&items).Error
	}
	return nil
}

const payslipSelect = `
SELECT ps.id, ps.period_id, ps.slip_no, ps.employee_id, ps.jam_mengajar, ps.jam_tidak_hadir, ps.hari_hadir,
       ps.total_earning, ps.total_deduction, ps.take_home_pay, ps.note,
       e.employee_no AS employee_no, e.full_name AS employee_name, e.job_title AS job_title,
       pp.effective_weeks AS effective_weeks
FROM payslips ps
JOIN employees e ON e.id = ps.employee_id
JOIN payroll_periods pp ON pp.id = ps.period_id `

type payslipRow struct {
	Payslip
	EmployeeNo     string   `gorm:"column:employee_no"`
	EmployeeName   string   `gorm:"column:employee_name"`
	JobTitle       *string  `gorm:"column:job_title"`
	EffectiveWeeks *float64 `gorm:"column:effective_weeks"`
}

func (r *SlipRepository) ListByPeriod(ctx context.Context, periodID int64, q string, limit, offset int64) ([]*payslipRow, int64, error) {
	where := "WHERE ps.period_id = ?"
	args := []any{periodID}
	if q != "" {
		where += " AND (e.full_name ILIKE ? OR e.employee_no ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like)
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM payslips ps JOIN employees e ON e.id = ps.employee_id "+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*payslipRow{}
	err := r.db.WithContext(ctx).Raw(
		payslipSelect+where+" ORDER BY ps.slip_no ASC LIMIT ? OFFSET ?",
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, total, err
}

func (r *SlipRepository) FindSlip(ctx context.Context, id int64) (*payslipRow, error) {
	row := &payslipRow{}
	err := r.db.WithContext(ctx).Raw(payslipSelect+"WHERE ps.id = ?", id).Scan(row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return row, nil
}

func (r *SlipRepository) ListItems(ctx context.Context, payslipID int64) ([]*PayslipItem, error) {
	items := []*PayslipItem{}
	err := r.db.WithContext(ctx).Where("payslip_id = ?", payslipID).
		Order("type ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *SlipRepository) FindItem(ctx context.Context, itemID int64) (*PayslipItem, error) {
	var item PayslipItem
	if err := r.db.WithContext(ctx).First(&item, itemID).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SlipRepository) CreateItem(ctx context.Context, item *PayslipItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *SlipRepository) UpdateItem(ctx context.Context, item *PayslipItem) error {
	return r.db.WithContext(ctx).Model(item).Select("amount", "note", "is_auto").Updates(item).Error
}

func (r *SlipRepository) DeleteItem(ctx context.Context, itemID int64) error {
	return r.db.WithContext(ctx).Delete(&PayslipItem{}, itemID).Error
}

func (r *SlipRepository) UpdateSlipNote(ctx context.Context, slipID int64, note *string) error {
	return r.db.WithContext(ctx).Exec("UPDATE payslips SET note = ? WHERE id = ?", note, slipID).Error
}

// RecomputeTotals menghitung ulang total earning/deduction/THP dari item slip.
func (r *SlipRepository) RecomputeTotals(ctx context.Context, tx *gorm.DB, payslipID int64) error {
	var totals struct {
		Earning   int64
		Deduction int64
	}
	err := tx.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE type = 'earning'), 0) AS earning,
			COALESCE(SUM(amount) FILTER (WHERE type = 'deduction'), 0) AS deduction
		FROM payslip_items WHERE payslip_id = ?
	`, payslipID).Scan(&totals).Error
	if err != nil {
		return err
	}
	return tx.WithContext(ctx).Exec(`
		UPDATE payslips SET total_earning = ?, total_deduction = ?, take_home_pay = ? WHERE id = ?
	`, totals.Earning, totals.Deduction, totals.Earning-totals.Deduction, payslipID).Error
}

func (r *SlipRepository) NextSlipNo(period *PayrollPeriod, seq int) string {
	return fmt.Sprintf("SLIP-%d%02d-%04d", period.PeriodYear, period.PeriodMonth, seq)
}

// UpdateAttendance menyimpan data jam (informasional) pada satu slip; kembali false bila slip bukan milik periode.
func (r *SlipRepository) UpdateAttendance(ctx context.Context, tx *gorm.DB, payslipID, periodID int64, jamMengajar, jamTidakHadir, hariHadir *int) (bool, error) {
	res := tx.WithContext(ctx).Exec(`
		UPDATE payslips SET jam_mengajar = ?, jam_tidak_hadir = ?, hari_hadir = ?
		WHERE id = ? AND period_id = ?
	`, jamMengajar, jamTidakHadir, hariHadir, payslipID, periodID)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
