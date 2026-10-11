package payroll

import (
	"time"

	"penggajian/server/internal/platform/httpx"
)

type ComponentResponse struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type SaveComponentRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type EmployeeComponentResponse struct {
	ID             int64     `json:"id"`
	EmployeeID     int64     `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	ComponentID    int64     `json:"component_id"`
	ComponentCode  string    `json:"component_code"`
	ComponentName  string    `json:"component_name"`
	ComponentType  string    `json:"component_type"`
	CalcMode       string    `json:"calc_mode"`
	Amount         int64     `json:"amount"`
	Note           *string   `json:"note"`
	CreatedAt      time.Time `json:"created_at"`
}

type SaveEmployeeComponentRequest struct {
	EmployeeID  int64   `json:"employee_id"`
	ComponentID int64   `json:"component_id"`
	CalcMode    string  `json:"calc_mode"`
	Amount      int64   `json:"amount"`
	Note        *string `json:"note"`
}

type PeriodResponse struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	PeriodMonth     int        `json:"period_month"`
	PeriodYear      int        `json:"period_year"`
	Status          string     `json:"status"`
	PaidAt          *time.Time `json:"paid_at"`
	Note            *string    `json:"note"`
	EffectiveWeeks  *float64   `json:"effective_weeks"`
	SlipCount       int64      `json:"slip_count"`
	TotalEarning    int64      `json:"total_earning"`
	TotalDeduction  int64      `json:"total_deduction"`
	TotalTakeHome   int64      `json:"total_take_home"`
	CreatedAt       time.Time  `json:"created_at"`
}

type SavePeriodRequest struct {
	Name        string  `json:"name"`
	PeriodMonth int     `json:"period_month"`
	PeriodYear  int     `json:"period_year"`
	Note        *string `json:"note"`
}

type PayslipResponse struct {
	ID             int64  `json:"id"`
	PeriodID       int64  `json:"period_id"`
	SlipNo         string `json:"slip_no"`
	EmployeeID     int64  `json:"employee_id"`
	EmployeeNo     string `json:"employee_no"`
	EmployeeName   string `json:"employee_name"`
	JobTitle       string `json:"job_title"`
	JamMengajar    *int   `json:"jam_mengajar"`
	JamTidakHadir  *int   `json:"jam_tidak_hadir"`
	HariHadir      *int   `json:"hari_hadir"`
	JmlHadir       *int   `json:"jml_hadir"`
	TotalEarning   int64  `json:"total_earning"`
	TotalDeduction int64  `json:"total_deduction"`
	TakeHomePay    int64  `json:"take_home_pay"`
	Note           *string `json:"note"`
}

type PayslipItemResponse struct {
	ID          int64   `json:"id"`
	ComponentID *int64  `json:"component_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Amount      int64   `json:"amount"`
	Note        *string `json:"note"`
	IsAuto      bool    `json:"is_auto"`
}

type PayslipDetailResponse struct {
	PayslipResponse
	Period *PeriodBrief `json:"period"`
	Items  []*PayslipItemResponse `json:"items"`
}

type PeriodBrief struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	PeriodMonth int    `json:"period_month"`
	PeriodYear  int    `json:"period_year"`
}

type SaveItemRequest struct {
	ComponentID *int64  `json:"component_id"`
	Name        *string `json:"name"`
	Type        *string `json:"type"`
	Amount      int64   `json:"amount"`
	Note        *string `json:"note"`
}

type UpdateItemRequest struct {
	Amount int64   `json:"amount"`
	Note   *string `json:"note"`
}

type GenerateResult struct {
	PeriodID  int64  `json:"period_id"`
	SlipCount int    `json:"slip_count"`
	Message   string `json:"message"`
}

type AttendanceItem struct {
	PayslipID     int64 `json:"payslip_id"`
	JamMengajar   *int  `json:"jam_mengajar"`
	JamTidakHadir *int  `json:"jam_tidak_hadir"`
	HariHadir     *int  `json:"hari_hadir"`
}

type BulkAttendanceRequest struct {
	Items []AttendanceItem `json:"items"`
}

type RecapRow struct {
	SlipNo         string           `json:"slip_no"`
	EmployeeNo     string           `json:"employee_no"`
	EmployeeName   string           `json:"employee_name"`
	Components     map[string]int64 `json:"components"`
	TotalEarning   int64            `json:"total_earning"`
	TotalDeduction int64            `json:"total_deduction"`
	TakeHomePay    int64            `json:"take_home_pay"`
}

type RecapResponse struct {
	Period          *PeriodResponse `json:"period"`
	Columns         []string        `json:"columns"`
	Rows            []*RecapRow     `json:"rows"`
	GrandEarning    int64           `json:"grand_earning"`
	GrandDeduction  int64           `json:"grand_deduction"`
	GrandTakeHome   int64           `json:"grand_take_home"`
}

type ComponentListResult struct {
	Components []*ComponentResponse `json:"components"`
	Meta       *httpx.Meta          `json:"meta"`
}

type EmployeeComponentListResult struct {
	EmployeeComponents []*EmployeeComponentResponse `json:"employee_components"`
	Meta               *httpx.Meta                  `json:"meta"`
}

type PeriodListResult struct {
	Periods []*PeriodResponse `json:"periods"`
	Meta    *httpx.Meta       `json:"meta"`
}

type PayslipListResult struct {
	Payslips []*PayslipResponse `json:"payslips"`
	Meta     *httpx.Meta        `json:"meta"`
}
