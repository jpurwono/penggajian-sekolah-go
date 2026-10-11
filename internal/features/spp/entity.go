package spp

import (
	"time"

	"gorm.io/gorm"
)

const (
	BillUnpaid  = "unpaid"
	BillPartial = "partial"
	BillPaid    = "paid"

	PayCash     = "cash"
	PayTransfer = "transfer"
	PayOther    = "lainnya"

	dateLayout = "2006-01-02"
)

var MonthNames = map[int]string{
	1: "Januari", 2: "Februari", 3: "Maret", 4: "April", 5: "Mei", 6: "Juni",
	7: "Juli", 8: "Agustus", 9: "September", 10: "Oktober", 11: "November", 12: "Desember",
}

type Student struct {
	ID           int64          `gorm:"primaryKey"`
	NIS          string         `gorm:"column:nis"`
	NISN         *string        `gorm:"column:nisn"`
	FullName     string         `gorm:"column:full_name"`
	ClassName    *string        `gorm:"column:class_name"`
	Gender       *string        `gorm:"column:gender"`
	GuardianName *string        `gorm:"column:guardian_name"`
	Phone        *string        `gorm:"column:phone"`
	Address      *string        `gorm:"column:address"`
	IsActive     bool           `gorm:"column:is_active"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Student) TableName() string { return "students" }

type Bill struct {
	ID          int64     `gorm:"primaryKey"`
	BillNo      string    `gorm:"column:bill_no"`
	StudentID   int64     `gorm:"column:student_id"`
	FeeTypeID   *int64    `gorm:"column:fee_type_id"`
	Name        string    `gorm:"column:name"`
	PeriodMonth *int      `gorm:"column:period_month"`
	PeriodYear  *int      `gorm:"column:period_year"`
	Amount      int64     `gorm:"column:amount"`
	PaidAmount  int64     `gorm:"column:paid_amount"`
	Status      string    `gorm:"column:status"`
	DueDate     *time.Time `gorm:"column:due_date"`
	Note        *string   `gorm:"column:note"`
	CreatedBy   *int64    `gorm:"column:created_by"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Bill) TableName() string { return "spp_bills" }

type Payment struct {
	ID         int64     `gorm:"primaryKey"`
	PaymentNo  string    `gorm:"column:payment_no"`
	BillID     int64     `gorm:"column:bill_id"`
	Amount     int64     `gorm:"column:amount"`
	PaidDate   time.Time `gorm:"column:paid_date"`
	Method     string    `gorm:"column:method"`
	Note       *string   `gorm:"column:note"`
	ReceivedBy *int64    `gorm:"column:received_by"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (Payment) TableName() string { return "spp_payments" }

func IsValidPaymentMethod(m string) bool {
	return m == PayCash || m == PayTransfer || m == PayOther
}
