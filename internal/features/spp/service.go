package spp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/middleware"
)

var ErrNotFound = errors.New("data tidak ditemukan")

type Service struct {
	db   *gorm.DB
	repo *Repository
}

func NewService(db *gorm.DB, repo *Repository) *Service {
	return &Service{db: db, repo: repo}
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func cleanStr(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func parseDate(v *string) (*time.Time, error) {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayoutStr, strings.TrimSpace(*v))
	if err != nil {
		return nil, errors.New("format tanggal harus YYYY-MM-DD")
	}
	return &t, nil
}

// ---------- Siswa ----------

func (s *Service) ListStudents(ctx context.Context, q, className string, active *bool, page, per int64) (*StudentListResult, error) {
	rows, total, err := s.repo.ListStudents(ctx, q, className, active, per, (page-1)*per)
	if err != nil {
		return nil, err
	}
	return &StudentListResult{Students: rows, Meta: httpx.NewMeta(page, per, total)}, nil
}

func (s *Service) validateStudent(ctx context.Context, req SaveStudentRequest, excludeID int64) (*Student, error) {
	nis := strings.TrimSpace(req.NIS)
	if nis == "" {
		return nil, errors.New("NIS wajib diisi")
	}
	if strings.TrimSpace(req.FullName) == "" {
		return nil, errors.New("nama siswa wajib diisi")
	}
	if exists, err := s.repo.NISExists(ctx, nis, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("NIS sudah digunakan")
	}
	gender := cleanStr(req.Gender)
	if gender != nil && (*gender != "L" && *gender != "P") {
		return nil, errors.New("jenis kelamin harus L atau P")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &Student{
		NIS: nis, NISN: cleanStr(req.NISN), FullName: strings.TrimSpace(req.FullName),
		ClassName: cleanStr(req.ClassName), Gender: gender,
		GuardianName: cleanStr(req.GuardianName), Phone: cleanStr(req.Phone),
		Address: cleanStr(req.Address), IsActive: isActive,
	}, nil
}

func (s *Service) CreateStudent(ctx context.Context, req SaveStudentRequest) (*StudentResponse, error) {
	student, err := s.validateStudent(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateStudent(ctx, student); err != nil {
		return nil, err
	}
	return &StudentResponse{
		ID: student.ID, NIS: student.NIS, NISN: student.NISN, FullName: student.FullName,
		ClassName: student.ClassName, Gender: student.Gender, GuardianName: student.GuardianName,
		Phone: student.Phone, Address: student.Address, IsActive: student.IsActive,
		CreatedAt: student.CreatedAt,
	}, nil
}

func (s *Service) UpdateStudent(ctx context.Context, id int64, req SaveStudentRequest) (*StudentResponse, error) {
	if _, err := s.repo.FindStudent(ctx, id); err != nil {
		return nil, mapNotFound(err)
	}
	student, err := s.validateStudent(ctx, req, id)
	if err != nil {
		return nil, err
	}
	student.ID = id
	if err := s.repo.UpdateStudent(ctx, student); err != nil {
		return nil, err
	}
	updated, err := s.repo.FindStudent(ctx, id)
	if err != nil {
		return nil, err
	}
	return &StudentResponse{
		ID: updated.ID, NIS: updated.NIS, NISN: updated.NISN, FullName: updated.FullName,
		ClassName: updated.ClassName, Gender: updated.Gender, GuardianName: updated.GuardianName,
		Phone: updated.Phone, Address: updated.Address, IsActive: updated.IsActive,
		CreatedAt: updated.CreatedAt,
	}, nil
}

func (s *Service) DeleteStudent(ctx context.Context, id int64) error {
	if _, err := s.repo.FindStudent(ctx, id); err != nil {
		return mapNotFound(err)
	}
	return s.repo.SoftDeleteStudent(ctx, id)
}

// ---------- Tagihan ----------

func (s *Service) ListBills(ctx context.Context, f BillFilter, page, per int64) (*BillListResult, *httpx.Meta, int64, int64, error) {
	rows, total, err := s.repo.ListBills(ctx, f, per, (page-1)*per)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	billed, paid, err := s.repo.BillTotals(ctx, f)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	return &BillListResult{Bills: rows, Meta: httpx.NewMeta(page, per, total)}, httpx.NewMeta(page, per, total), billed, paid, nil
}

func (s *Service) GetBill(ctx context.Context, id int64) (*BillResponse, error) {
	row, err := s.repo.FindBill(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return row, nil
}

func (s *Service) validateBill(ctx context.Context, req SaveBillRequest) (*Bill, error) {
	if req.StudentID <= 0 {
		return nil, errors.New("siswa wajib dipilih")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("nama tagihan wajib diisi")
	}
	if req.Amount <= 0 {
		return nil, errors.New("nominal tagihan harus lebih dari 0")
	}
	if req.PeriodMonth != nil && (*req.PeriodMonth < 1 || *req.PeriodMonth > 12) {
		return nil, errors.New("bulan harus 1-12")
	}
	if _, err := s.repo.FindStudent(ctx, req.StudentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("siswa tidak ditemukan")
		}
		return nil, err
	}
	dueDate, err := parseDate(req.DueDate)
	if err != nil {
		return nil, err
	}
	return &Bill{
		StudentID: req.StudentID, Name: strings.TrimSpace(req.Name),
		PeriodMonth: req.PeriodMonth, PeriodYear: req.PeriodYear,
		Amount: req.Amount, DueDate: dueDate, Note: cleanStr(req.Note),
		Status: BillUnpaid,
	}, nil
}

func (s *Service) CreateBill(ctx context.Context, c echo.Context, req SaveBillRequest) (*BillResponse, error) {
	bill, err := s.validateBill(ctx, req)
	if err != nil {
		return nil, err
	}
	actorID := middleware.CurrentUserID(c)
	bill.CreatedBy = &actorID

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		billNo, err := s.repo.NextBillNo(ctx, tx, time.Now())
		if err != nil {
			return err
		}
		bill.BillNo = billNo
		return s.repo.CreateBill(ctx, tx, bill)
	})
	if err != nil {
		return nil, err
	}
	audit.Log(c, s.db, "create", "spp_bills", bill.ID, nil, bill)
	return s.GetBill(ctx, bill.ID)
}

func (s *Service) UpdateBill(ctx context.Context, c echo.Context, id int64, req SaveBillRequest) (*BillResponse, error) {
	existing, err := s.repo.FindBillEntity(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if existing.PaidAmount > 0 {
		return nil, errors.New("tagihan yang sudah memiliki pembayaran tidak dapat diubah")
	}
	bill, err := s.validateBill(ctx, req)
	if err != nil {
		return nil, err
	}
	bill.ID = id
	if err := s.repo.UpdateBill(ctx, bill); err != nil {
		return nil, err
	}
	audit.Log(c, s.db, "update", "spp_bills", id, nil, bill)
	return s.GetBill(ctx, id)
}

func (s *Service) DeleteBill(ctx context.Context, c echo.Context, id int64) error {
	existing, err := s.repo.FindBillEntity(ctx, id)
	if err != nil {
		return mapNotFound(err)
	}
	if existing.PaidAmount > 0 {
		return errors.New("tagihan yang sudah dibayar tidak dapat dihapus")
	}
	if err := s.repo.DeleteBill(ctx, id); err != nil {
		return err
	}
	audit.Log(c, s.db, "delete", "spp_bills", id, nil, nil)
	return nil
}

// Generate membuat tagihan untuk seluruh siswa aktif pada bulan/tahun tertentu.
func (s *Service) Generate(ctx context.Context, c echo.Context, req GenerateBillsRequest) (*GenerateResult, error) {
	if req.PeriodMonth < 1 || req.PeriodMonth > 12 {
		return nil, errors.New("bulan harus 1-12")
	}
	if req.PeriodYear < 1900 || req.PeriodYear > 2100 {
		return nil, errors.New("tahun tidak valid")
	}
	if req.Amount <= 0 {
		return nil, errors.New("nominal tagihan harus lebih dari 0")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("SPP %s %d", MonthNames[req.PeriodMonth], req.PeriodYear)
	}

	className := ""
	if req.ClassName != nil {
		className = strings.TrimSpace(*req.ClassName)
	}
	students, err := s.repo.ListActiveStudents(ctx, className)
	if err != nil {
		return nil, err
	}
	if len(students) == 0 {
		return nil, errors.New("belum ada siswa aktif untuk digenerate")
	}

	dueDate, err := parseDate(req.DueDate)
	if err != nil {
		return nil, err
	}

	actorID := middleware.CurrentUserID(c)
	created, skipped := 0, 0

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, student := range students {
			exists, err := s.repo.BillExistsForStudent(ctx, student.ID, name, req.PeriodMonth, req.PeriodYear)
			if err != nil {
				return err
			}
			if exists {
				skipped++
				continue
			}
			billNo, err := s.repo.NextBillNo(ctx, tx, time.Now())
			if err != nil {
				return err
			}
			month, year := req.PeriodMonth, req.PeriodYear
			bill := &Bill{
				BillNo: billNo, StudentID: student.ID, Name: name,
				PeriodMonth: &month, PeriodYear: &year,
				Amount: req.Amount, DueDate: dueDate, CreatedBy: &actorID,
				Status: BillUnpaid,
			}
			if err := s.repo.CreateBill(ctx, tx, bill); err != nil {
				return err
			}
			created++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	audit.Log(c, s.db, "generate", "spp_bills", 0, nil, map[string]any{
		"name": name, "created": created, "skipped": skipped,
	})
	return &GenerateResult{
		Created: created, Skipped: skipped,
		Message: fmt.Sprintf("%d tagihan dibuat, %d dilewati (sudah ada)", created, skipped),
	}, nil
}

// ---------- Pembayaran ----------

func (s *Service) CreatePayment(ctx context.Context, c echo.Context, billID int64, req CreatePaymentRequest) (*PaymentDetailResponse, error) {
	bill, err := s.repo.FindBillEntity(ctx, billID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	remaining := bill.Amount - bill.PaidAmount
	if remaining <= 0 {
		return nil, errors.New("tagihan sudah lunas")
	}
	if req.Amount <= 0 {
		return nil, errors.New("nominal pembayaran harus lebih dari 0")
	}
	if req.Amount > remaining {
		return nil, fmt.Errorf("nominal melebihi sisa tagihan (sisa %d)", remaining)
	}
	if !IsValidPaymentMethod(req.Method) {
		return nil, errors.New("metode pembayaran tidak valid (cash/transfer/lainnya)")
	}

	paidDate := time.Now()
	if strings.TrimSpace(req.PaidDate) != "" {
		paidDate, err = time.Parse(dateLayoutStr, strings.TrimSpace(req.PaidDate))
		if err != nil {
			return nil, errors.New("format tanggal bayar harus YYYY-MM-DD")
		}
	}

	actorID := middleware.CurrentUserID(c)
	payment := &Payment{
		BillID: billID, Amount: req.Amount, PaidDate: paidDate,
		Method: req.Method, Note: cleanStr(req.Note), ReceivedBy: &actorID,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		paymentNo, err := s.repo.NextPaymentNo(ctx, tx, paidDate)
		if err != nil {
			return err
		}
		payment.PaymentNo = paymentNo
		if err := s.repo.CreatePayment(ctx, tx, payment); err != nil {
			return err
		}

		newPaid := bill.PaidAmount + req.Amount
		status := BillPartial
		if newPaid >= bill.Amount {
			status = BillPaid
		}
		return tx.Exec(`UPDATE spp_bills SET paid_amount = ?, status = ? WHERE id = ?`, newPaid, status, billID).Error
	})
	if err != nil {
		return nil, err
	}

	audit.Log(c, s.db, "pay", "spp_payments", payment.ID, nil, map[string]any{
		"payment_no": payment.PaymentNo, "bill_no": bill.BillNo, "amount": req.Amount,
	})

	detail, err := s.GetPayment(ctx, payment.ID)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (s *Service) ListPayments(ctx context.Context, f PaymentFilter, page, per int64) (*PaymentListResult, error) {
	rows, total, err := s.repo.ListPayments(ctx, f, per, (page-1)*per)
	if err != nil {
		return nil, err
	}
	return &PaymentListResult{Payments: rows, Meta: httpx.NewMeta(page, per, total)}, nil
}

func (s *Service) GetPayment(ctx context.Context, id int64) (*PaymentDetailResponse, error) {
	row, err := s.repo.FindPayment(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	bill, err := s.GetBill(ctx, row.BillID)
	if err != nil {
		return nil, err
	}
	return &PaymentDetailResponse{PaymentResponse: *row, Bill: bill}, nil
}

func (s *Service) ListBillPayments(ctx context.Context, billID int64) ([]*PaymentResponse, error) {
	return s.repo.ListBillPayments(ctx, billID)
}

func (s *Service) YearSummary(ctx context.Context, year int) (*YearSummary, error) {
	return s.repo.YearSummary(ctx, year)
}
