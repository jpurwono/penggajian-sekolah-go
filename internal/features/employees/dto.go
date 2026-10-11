package employees

import (
	"time"

	"penggajian/server/internal/platform/httpx"
)

const dateLayout = "2006-01-02"

type EmployeeResponse struct {
	ID           int64     `json:"id"`
	EmployeeNo   string    `json:"employee_no"`
	FullName     string    `json:"full_name"`
	Gender       *string   `json:"gender"`
	BirthDate    *string   `json:"birth_date"`
	Address      *string   `json:"address"`
	Phone        *string   `json:"phone"`
	Email        *string   `json:"email"`
	EmployeeType string    `json:"employee_type"`
	JobTitle     *string   `json:"job_title"`
	JoinDate     string    `json:"join_date"`
	EndDate      *string   `json:"end_date"`
	BaseSalary   int64     `json:"base_salary"`
	BankName     *string   `json:"bank_name"`
	BankAccount  *string   `json:"bank_account"`
	PhotoURL     *string   `json:"photo_url"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type EmployeeListResult struct {
	Employees []*EmployeeResponse `json:"employees"`
	Meta      *httpx.Meta         `json:"meta"`
}

type SaveEmployeeRequest struct {
	EmployeeNo   string  `json:"employee_no"`
	FullName     string  `json:"full_name"`
	Gender       *string `json:"gender"`
	BirthDate    *string `json:"birth_date"`
	Address      *string `json:"address"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
	EmployeeType string  `json:"employee_type"`
	JobTitle     *string `json:"job_title"`
	JoinDate     string  `json:"join_date"`
	EndDate      *string `json:"end_date"`
	BaseSalary   int64   `json:"base_salary"`
	BankName     *string `json:"bank_name"`
	BankAccount  *string `json:"bank_account"`
	PhotoURL     *string `json:"photo_url"`
	IsActive     *bool   `json:"is_active"`
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(dateLayout)
	return &s
}

func toResponse(e *Employee) *EmployeeResponse {
	return &EmployeeResponse{
		ID:           e.ID,
		EmployeeNo:   e.EmployeeNo,
		FullName:     e.FullName,
		Gender:       e.Gender,
		BirthDate:    formatDate(e.BirthDate),
		Address:      e.Address,
		Phone:        e.Phone,
		Email:        e.Email,
		EmployeeType: e.EmployeeType,
		JobTitle:     e.JobTitle,
		JoinDate:     e.JoinDate.Format(dateLayout),
		EndDate:      formatDate(e.EndDate),
		BaseSalary:   e.BaseSalary,
		BankName:     e.BankName,
		BankAccount:  e.BankAccount,
		PhotoURL:     e.PhotoURL,
		IsActive:     e.IsActive,
		CreatedAt:    e.CreatedAt,
	}
}
