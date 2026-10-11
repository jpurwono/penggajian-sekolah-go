package employees

import (
	"time"

	"gorm.io/gorm"
)

type Employee struct {
	ID           int64          `gorm:"primaryKey"`
	EmployeeNo   string         `gorm:"column:employee_no"`
	FullName     string         `gorm:"column:full_name"`
	Gender       *string        `gorm:"column:gender"`
	BirthDate    *time.Time     `gorm:"column:birth_date"`
	Address      *string        `gorm:"column:address"`
	Phone        *string        `gorm:"column:phone"`
	Email        *string        `gorm:"column:email"`
	EmployeeType string         `gorm:"column:employee_type"`
	JobTitle     *string        `gorm:"column:job_title"`
	JoinDate     time.Time      `gorm:"column:join_date"`
	EndDate      *time.Time     `gorm:"column:end_date"`
	BaseSalary   int64          `gorm:"column:base_salary"`
	BankName     *string        `gorm:"column:bank_name"`
	BankAccount  *string        `gorm:"column:bank_account"`
	PhotoURL     *string        `gorm:"column:photo_url"`
	IsActive     bool           `gorm:"column:is_active"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Employee) TableName() string { return "employees" }

var validEmployeeTypes = map[string]bool{
	"tetap":    true,
	"pppk":     true,
	"honorer":  true,
	"staff_tu": true,
}

func IsValidEmployeeType(t string) bool { return validEmployeeTypes[t] }
