package spp

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

// ---------- Siswa ----------

func (r *Repository) ListStudents(ctx context.Context, q, className string, active *bool, limit, offset int64) ([]*StudentResponse, int64, error) {
	where := "WHERE s.deleted_at IS NULL"
	args := []any{}
	if q != "" {
		where += " AND (s.nis ILIKE ? OR COALESCE(s.nisn, '') ILIKE ? OR s.full_name ILIKE ? OR COALESCE(s.guardian_name, '') ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like, like)
	}
	if className != "" {
		where += " AND s.class_name = ?"
		args = append(args, className)
	}
	if active != nil {
		where += " AND s.is_active = ?"
		args = append(args, *active)
	}

	var total int64
	if err := r.dbCtx(ctx).Raw("SELECT COUNT(*) FROM students s "+where, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*StudentResponse{}
	err := r.dbCtx(ctx).Raw(`
		SELECT s.id, s.nis, s.nisn, s.full_name, s.class_name, s.gender,
		       s.guardian_name, s.phone, s.address, s.is_active, s.created_at
		FROM students s
		`+where+` ORDER BY s.class_name ASC NULLS LAST, s.full_name ASC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) FindStudent(ctx context.Context, id int64) (*Student, error) {
	var s Student
	if err := r.dbCtx(ctx).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) CreateStudent(ctx context.Context, s *Student) error {
	return r.dbCtx(ctx).Create(s).Error
}

func (r *Repository) UpdateStudent(ctx context.Context, s *Student) error {
	return r.dbCtx(ctx).Model(s).
		Select("nis", "nisn", "full_name", "class_name", "gender", "guardian_name", "phone", "address", "is_active").
		Updates(s).Error
}

func (r *Repository) SoftDeleteStudent(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&Student{}, id).Error
}

func (r *Repository) NISExists(ctx context.Context, nis string, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Unscoped().Model(&Student{}).
		Where("nis = ? AND id <> ?", nis, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListActiveStudents(ctx context.Context, className string) ([]*Student, error) {
	query := r.dbCtx(ctx).Model(&Student{}).Where("is_active = ?", true)
	if className != "" {
		query = query.Where("class_name = ?", className)
	}
	var list []*Student
	err := query.Order("class_name ASC NULLS LAST, full_name ASC").Find(&list).Error
	return list, err
}

// ---------- Tagihan ----------

const billSelect = `
SELECT b.id, b.bill_no, b.student_id, s.nis AS student_nis, s.full_name AS student_name,
       s.class_name, b.name, b.period_month, b.period_year,
       b.amount, b.paid_amount, (b.amount - b.paid_amount) AS remaining,
       b.status, b.due_date, b.note, b.created_at
FROM spp_bills b
JOIN students s ON s.id = b.student_id `

type BillFilter struct {
	Q            string
	StudentID    *int64
	PeriodMonth  int
	PeriodYear   int
	Status       string
	ClassName    string
}

func billFilterWhere(f BillFilter) (string, []any) {
	where := "WHERE 1=1"
	args := []any{}
	if f.Q != "" {
		where += " AND (b.name ILIKE ? OR s.full_name ILIKE ? OR s.nis ILIKE ? OR b.bill_no ILIKE ?)"
		like := "%" + f.Q + "%"
		args = append(args, like, like, like, like)
	}
	if f.StudentID != nil {
		where += " AND b.student_id = ?"
		args = append(args, *f.StudentID)
	}
	if f.PeriodMonth > 0 {
		where += " AND b.period_month = ?"
		args = append(args, f.PeriodMonth)
	}
	if f.PeriodYear > 0 {
		where += " AND b.period_year = ?"
		args = append(args, f.PeriodYear)
	}
	if f.Status != "" {
		where += " AND b.status = ?"
		args = append(args, f.Status)
	}
	if f.ClassName != "" {
		where += " AND s.class_name = ?"
		args = append(args, f.ClassName)
	}
	return where, args
}

func (r *Repository) ListBills(ctx context.Context, f BillFilter, limit, offset int64) ([]*BillResponse, int64, error) {
	where, args := billFilterWhere(f)

	var total int64
	if err := r.dbCtx(ctx).Raw(
		"SELECT COUNT(*) FROM spp_bills b JOIN students s ON s.id = b.student_id "+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*BillResponse{}
	err := r.dbCtx(ctx).Raw(
		billSelect+where+" ORDER BY b.id DESC LIMIT ? OFFSET ?",
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) BillTotals(ctx context.Context, f BillFilter) (billed, paid int64, err error) {
	where, args := billFilterWhere(f)
	err = r.dbCtx(ctx).Raw(`
		SELECT COALESCE(SUM(b.amount), 0) AS billed, COALESCE(SUM(b.paid_amount), 0) AS paid
		FROM spp_bills b JOIN students s ON s.id = b.student_id `+where,
		args...,
	).Row().Scan(&billed, &paid)
	return
}

func (r *Repository) FindBill(ctx context.Context, id int64) (*BillResponse, error) {
	rows := []*BillResponse{}
	err := r.dbCtx(ctx).Raw(billSelect+"WHERE b.id = ?", id).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return rows[0], nil
}

func (r *Repository) FindBillEntity(ctx context.Context, id int64) (*Bill, error) {
	var b Bill
	if err := r.dbCtx(ctx).First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) CreateBill(ctx context.Context, tx *gorm.DB, b *Bill) error {
	return tx.WithContext(ctx).Create(b).Error
}

func (r *Repository) UpdateBill(ctx context.Context, b *Bill) error {
	return r.dbCtx(ctx).Model(b).
		Select("name", "period_month", "period_year", "amount", "due_date", "note").Updates(b).Error
}

func (r *Repository) DeleteBill(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&Bill{}, id).Error
}

func (r *Repository) BillExistsForStudent(ctx context.Context, studentID int64, name string, month, year int) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&Bill{}).
		Where("student_id = ? AND name = ? AND period_month = ? AND period_year = ?", studentID, name, month, year).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) NextBillNo(ctx context.Context, tx *gorm.DB, date time.Time) (string, error) {
	prefix := fmt.Sprintf("TAG-%d%02d-", date.Year(), int(date.Month()))
	var count int64
	if err := tx.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM spp_bills WHERE bill_no LIKE ?", prefix+"%",
	).Scan(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *Repository) NextPaymentNo(ctx context.Context, tx *gorm.DB, date time.Time) (string, error) {
	prefix := fmt.Sprintf("PAY-%d%02d-", date.Year(), int(date.Month()))
	var count int64
	if err := tx.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM spp_payments WHERE payment_no LIKE ?", prefix+"%",
	).Scan(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

// ---------- Pembayaran ----------

const paymentSelect = `
SELECT p.id, p.payment_no, p.bill_id, b.bill_no, b.name AS bill_name,
       s.id AS student_id, s.nis AS student_nis, s.full_name AS student_name, s.class_name,
       p.amount, p.paid_date, p.method, p.note, p.received_by, p.created_at
FROM spp_payments p
JOIN spp_bills b ON b.id = p.bill_id
JOIN students s ON s.id = b.student_id `

type PaymentFilter struct {
	Q        string
	StudentID *int64
	DateFrom string
	DateTo   string
}

func paymentFilterWhere(f PaymentFilter) (string, []any) {
	where := "WHERE 1=1"
	args := []any{}
	if f.Q != "" {
		where += " AND (p.payment_no ILIKE ? OR s.full_name ILIKE ? OR s.nis ILIKE ? OR b.bill_no ILIKE ?)"
		like := "%" + f.Q + "%"
		args = append(args, like, like, like, like)
	}
	if f.StudentID != nil {
		where += " AND b.student_id = ?"
		args = append(args, *f.StudentID)
	}
	if f.DateFrom != "" {
		where += " AND p.paid_date >= ?"
		args = append(args, f.DateFrom)
	}
	if f.DateTo != "" {
		where += " AND p.paid_date <= ?"
		args = append(args, f.DateTo)
	}
	return where, args
}

func (r *Repository) ListPayments(ctx context.Context, f PaymentFilter, limit, offset int64) ([]*PaymentResponse, int64, error) {
	where, args := paymentFilterWhere(f)

	var total int64
	if err := r.dbCtx(ctx).Raw(`
		SELECT COUNT(*) FROM spp_payments p
		JOIN spp_bills b ON b.id = p.bill_id
		JOIN students s ON s.id = b.student_id `+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*PaymentResponse{}
	err := r.dbCtx(ctx).Raw(
		paymentSelect+where+" ORDER BY p.id DESC LIMIT ? OFFSET ?",
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) FindPayment(ctx context.Context, id int64) (*PaymentResponse, error) {
	rows := []*PaymentResponse{}
	err := r.dbCtx(ctx).Raw(paymentSelect+"WHERE p.id = ?", id).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return rows[0], nil
}

func (r *Repository) CreatePayment(ctx context.Context, tx *gorm.DB, p *Payment) error {
	return tx.WithContext(ctx).Create(p).Error
}

func (r *Repository) ListBillPayments(ctx context.Context, billID int64) ([]*PaymentResponse, error) {
	rows := []*PaymentResponse{}
	err := r.dbCtx(ctx).Raw(paymentSelect+"WHERE p.bill_id = ? ORDER BY p.id ASC", billID).Scan(&rows).Error
	return rows, err
}

// ---------- Ringkasan ----------

func (r *Repository) YearSummary(ctx context.Context, year int) (*YearSummary, error) {
	rows := []*MonthSummaryRow{}
	err := r.dbCtx(ctx).Raw(`
		SELECT COALESCE(b.period_month, 0) AS month,
		       COUNT(*) AS bill_count,
		       COALESCE(SUM(b.amount), 0) AS billed,
		       COALESCE(SUM(b.paid_amount), 0) AS collected,
		       COALESCE(SUM(b.amount - b.paid_amount), 0) AS outstanding
		FROM spp_bills b
		WHERE b.period_year = ?
		GROUP BY 1
		ORDER BY 1
	`, year).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	totals := &MonthSummaryRow{}
	for _, row := range rows {
		totals.BillCount += row.BillCount
		totals.Billed += row.Billed
		totals.Collected += row.Collected
		totals.Outstanding += row.Outstanding
	}
	return &YearSummary{Year: year, Months: rows, Totals: totals}, nil
}
