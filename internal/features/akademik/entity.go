package akademik

import (
	"time"

	"gorm.io/gorm"
)

type Subject struct {
	ID        int64          `gorm:"primaryKey"`
	Code      string         `gorm:"column:code"`
	Name      string         `gorm:"column:name"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Subject) TableName() string { return "subjects" }

type SchoolClass struct {
	ID          int64          `gorm:"primaryKey"`
	Name        string         `gorm:"column:name"`
	Level       *string        `gorm:"column:level"`
	Description *string        `gorm:"column:description"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (SchoolClass) TableName() string { return "school_classes" }

type TeachingAssignment struct {
	ID         int64     `gorm:"primaryKey"`
	EmployeeID int64     `gorm:"column:employee_id"`
	SubjectID  int64     `gorm:"column:subject_id"`
	ClassID    int64     `gorm:"column:class_id"`
	DayOfWeek  int       `gorm:"column:day_of_week"`
	StartTime  string    `gorm:"column:start_time"` // HH:MM
	EndTime    string    `gorm:"column:end_time"`   // HH:MM
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (TeachingAssignment) TableName() string { return "teaching_assignments" }

var DayNames = map[int]string{
	1: "Senin", 2: "Selasa", 3: "Rabu", 4: "Kamis", 5: "Jumat", 6: "Sabtu", 7: "Minggu",
}

func IsValidDay(d int) bool { return d >= 1 && d <= 7 }
