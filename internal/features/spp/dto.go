package spp

import (
	"time"

	"penggajian/server/internal/platform/httpx"
)

const dateLayoutStr = "2006-01-02"

type StudentResponse struct {
	ID           int64     `json:"id"`
	NIS          string    `json:"nis"`
	NISN         *string   `json:"nisn"`
	FullName     string    `json:"full_name"`
	ClassName    *string   `json:"class_name"`
	Gender       *string   `json:"gender"`
	GuardianName *string   `json:"guardian_name"`
	Phone        *string   `json:"phone"`
	Address      *string   `json:"address"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type SaveStudentRequest struct {
	NIS          string  `json:"nis"`
	NISN         *string `json:"nisn"`
	FullName     string  `json:"full_name"`
	ClassName    *string `json:"class_name"`
	Gender       *string `json:"gender"`
	GuardianName *string `json:"guardian_name"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	IsActive     *bool   `json:"is_active"`
}

type BillResponse struct {
	ID          int64     `json:"id"`
	BillNo      string    `json:"bill_no"`
	StudentID   int64     `json:"student_id"`
	StudentNIS  string    `json:"student_nis"`
	StudentName string    `json:"student_name"`
	ClassName   *string   `json:"class_name"`
	Name        string    `json:"name"`
	PeriodMonth *int      `json:"period_month"`
	PeriodYear  *int      `json:"period_year"`
	Amount      int64     `json:"amount"`
	PaidAmount  int64     `json:"paid_amount"`
	Remaining   int64     `json:"remaining"`
	Status      string    `json:"status"`
	DueDate     *string   `json:"due_date"`
	Note        *string   `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
}

type SaveBillRequest struct {
	StudentID   int64   `json:"student_id"`
	Name        string  `json:"name"`
	PeriodMonth *int    `json:"period_month"`
	PeriodYear  *int    `json:"period_year"`
	Amount      int64   `json:"amount"`
	DueDate     *string `json:"due_date"`
	Note        *string `json:"note"`
}

type GenerateBillsRequest struct {
	Name        string  `json:"name"`
	PeriodMonth int     `json:"period_month"`
	PeriodYear  int     `json:"period_year"`
	Amount      int64   `json:"amount"`
	DueDate     *string `json:"due_date"`
	ClassName   *string `json:"class_name"`
}

type GenerateResult struct {
	Created int    `json:"created"`
	Skipped int    `json:"skipped"`
	Message string `json:"message"`
}

type PaymentResponse struct {
	ID          int64     `json:"id"`
	PaymentNo   string    `json:"payment_no"`
	BillID      int64     `json:"bill_id"`
	BillNo      string    `json:"bill_no"`
	BillName    string    `json:"bill_name"`
	StudentID   int64     `json:"student_id"`
	StudentNIS  string    `json:"student_nis"`
	StudentName string    `json:"student_name"`
	ClassName   *string   `json:"class_name"`
	Amount      int64     `json:"amount"`
	PaidDate    string    `json:"paid_date"`
	Method      string    `json:"method"`
	Note        *string   `json:"note"`
	ReceivedBy  *int64    `json:"received_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreatePaymentRequest struct {
	Amount   int64   `json:"amount"`
	PaidDate string  `json:"paid_date"`
	Method   string  `json:"method"`
	Note     *string `json:"note"`
}

type PaymentDetailResponse struct {
	PaymentResponse
	Bill *BillResponse `json:"bill"`
}

type MonthSummaryRow struct {
	Month      int    `json:"month"`
	BillCount  int64  `json:"bill_count"`
	Billed     int64  `json:"billed"`
	Collected  int64  `json:"collected"`
	Outstanding int64 `json:"outstanding"`
}

type YearSummary struct {
	Year    int                `json:"year"`
	Months  []*MonthSummaryRow `json:"months"`
	Totals  *MonthSummaryRow   `json:"totals"`
}

type StudentListResult struct {
	Students []*StudentResponse `json:"students"`
	Meta     *httpx.Meta        `json:"meta"`
}

type BillListResult struct {
	Bills []*BillResponse `json:"bills"`
	Meta  *httpx.Meta     `json:"meta"`
}

type PaymentListResult struct {
	Payments []*PaymentResponse `json:"payments"`
	Meta     *httpx.Meta        `json:"meta"`
}

func formatBillDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(dateLayoutStr)
	return &s
}
