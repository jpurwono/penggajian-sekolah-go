package attendance

import "time"

type AttendanceResponse struct {
	ID           int64      `json:"id"`
	EmployeeID   int64      `json:"employee_id"`
	EmployeeName string     `json:"employee_name"`
	EmployeeNo   string     `json:"employee_no"`
	Jenis        string     `json:"jenis"`
	JenisLabel   string     `json:"jenis_label"`
	TeachingID   *int64     `json:"teaching_id"`
	ClassName    *string    `json:"class_name,omitempty"`
	SubjectName  *string    `json:"subject_name,omitempty"`
	DayName      *string    `json:"day_name,omitempty"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
	AccuracyM    *float64   `json:"accuracy_m"`
	IsMock       bool       `json:"is_mock"`
	DistanceM    *float64   `json:"distance_m"`
	RecordedAt   time.Time  `json:"recorded_at"`
	HasPhoto     bool       `json:"has_photo"`
}

type CheckInRequest struct {
	Jenis      string  `json:"jenis"`
	TeachingID *int64  `json:"teaching_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	AccuracyM  *float64 `json:"accuracy_m"`
	IsMock     bool    `json:"is_mock"`
	RecordedAt string  `json:"recorded_at"`
	DeviceInfo string  `json:"device_info"`
}

type CheckInResult struct {
	Record   *AttendanceResponse `json:"record"`
	Distance float64             `json:"distance_m"`
	Within   bool                `json:"within_geofence"`
	Message  string              `json:"message"`
}

type DailySummaryRow struct {
	EmployeeID   int64  `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	EmployeeNo   string `json:"employee_no"`
	HadirCount   int    `json:"hadir_count"`
	KelasCount   int    `json:"kelas_count"`
	LastHadir    *time.Time `json:"last_hadir"`
	LastKelas    *time.Time `json:"last_kelas"`
}

type DailySummary struct {
	Date string              `json:"date"`
	Rows []*DailySummaryRow  `json:"rows"`
}
