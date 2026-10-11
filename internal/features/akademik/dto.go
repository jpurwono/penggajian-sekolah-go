package akademik

import (
	"time"

	"penggajian/server/internal/platform/httpx"
)

type SubjectResponse struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type SaveSubjectRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ClassResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Level       *string   `json:"level"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type SaveClassRequest struct {
	Name        string  `json:"name"`
	Level       *string `json:"level"`
	Description *string `json:"description"`
}

type TeachingResponse struct {
	ID         int64  `json:"id"`
	EmployeeID int64  `json:"employee_id"`
	SubjectID  int64  `json:"subject_id"`
	SubjectName string `json:"subject_name"`
	ClassID    int64  `json:"class_id"`
	ClassName  string `json:"class_name"`
	DayOfWeek  int    `json:"day_of_week"`
	DayName    string `json:"day_name"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}

type SaveTeachingRequest struct {
	SubjectID int64  `json:"subject_id"`
	ClassID   int64  `json:"class_id"`
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type SubjectListResult struct {
	Subjects []*SubjectResponse `json:"subjects"`
	Meta     *httpx.Meta        `json:"meta"`
}

type ClassListResult struct {
	Classes []*ClassResponse `json:"classes"`
	Meta    *httpx.Meta      `json:"meta"`
}

type TeachingListResult struct {
	Teachings []*TeachingResponse `json:"teachings"`
	Meta      *httpx.Meta         `json:"meta"`
}
