package payroll

import (
	"time"

	"gorm.io/gorm"
)

const (
	TypeEarning   = "earning"
	TypeDeduction = "deduction"
)

const (
	CalcFixed          = "fixed"
	CalcPerJamHadir    = "per_jam_hadir"
	CalcPerJamMengajar = "per_jam_mengajar"
	CalcPerHariHadir   = "per_hari_hadir"
	CalcPerJamMinggu1  = "per_jam_minggu1"
)

func IsValidCalcMode(m string) bool {
	return m == CalcFixed || m == CalcPerJamHadir || m == CalcPerJamMengajar ||
		m == CalcPerHariHadir || m == CalcPerJamMinggu1
}

const ComponentCodeGajiPokok = "GP"

type PayComponent struct {
	ID          int64          `gorm:"primaryKey"`
	Code        string         `gorm:"column:code"`
	Name        string         `gorm:"column:name"`
	Type        string         `gorm:"column:type"`
	Description *string        `gorm:"column:description"`
	IsActive    bool           `gorm:"column:is_active"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (PayComponent) TableName() string { return "pay_components" }

type EmployeeComponent struct {
	ID           int64     `gorm:"primaryKey"`
	EmployeeID   int64     `gorm:"column:employee_id"`
	ComponentID  int64     `gorm:"column:component_id"`
	CalcMode     string    `gorm:"column:calc_mode"`
	Amount       int64     `gorm:"column:amount"`
	Note         *string   `gorm:"column:note"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (EmployeeComponent) TableName() string { return "employee_components" }

type PayrollPeriod struct {
	ID             int64          `gorm:"primaryKey"`
	Name           string         `gorm:"column:name"`
	PeriodMonth    int            `gorm:"column:period_month"`
	PeriodYear     int            `gorm:"column:period_year"`
	Status         string         `gorm:"column:status"`
	PaidAt         *time.Time     `gorm:"column:paid_at"`
	Note           *string        `gorm:"column:note"`
	EffectiveWeeks *float64       `gorm:"column:effective_weeks"`
	CreatedBy      *int64         `gorm:"column:created_by"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (PayrollPeriod) TableName() string { return "payroll_periods" }

const (
	StatusDraft     = "draft"
	StatusApproved  = "approved"
	StatusPaid      = "paid"
	StatusCancelled = "cancelled"
)

type Payslip struct {
	ID             int64     `gorm:"primaryKey"`
	PeriodID       int64     `gorm:"column:period_id"`
	EmployeeID     int64     `gorm:"column:employee_id"`
	SlipNo         string    `gorm:"column:slip_no"`
	JamMengajar    *int      `gorm:"column:jam_mengajar"`
	JamTidakHadir  *int      `gorm:"column:jam_tidak_hadir"`
	HariHadir      *int      `gorm:"column:hari_hadir"`
	TotalEarning   int64     `gorm:"column:total_earning"`
	TotalDeduction int64     `gorm:"column:total_deduction"`
	TakeHomePay    int64     `gorm:"column:take_home_pay"`
	Note           *string   `gorm:"column:note"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (Payslip) TableName() string { return "payslips" }

type PayslipItem struct {
	ID           int64     `gorm:"primaryKey"`
	PayslipID    int64     `gorm:"column:payslip_id"`
	ComponentID  *int64    `gorm:"column:component_id"`
	CodeSnapshot string    `gorm:"column:code_snapshot"`
	NameSnapshot string    `gorm:"column:name_snapshot"`
	Type         string    `gorm:"column:type"`
	Amount       int64     `gorm:"column:amount"`
	Note         *string   `gorm:"column:note"`
	IsAuto       bool      `gorm:"column:is_auto"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (PayslipItem) TableName() string { return "payslip_items" }

var MonthNames = map[int]string{
	1: "Januari", 2: "Februari", 3: "Maret", 4: "April", 5: "Mei", 6: "Juni",
	7: "Juli", 8: "Agustus", 9: "September", 10: "Oktober", 11: "November", 12: "Desember",
}

func IsValidComponentType(t string) bool { return t == TypeEarning || t == TypeDeduction }
