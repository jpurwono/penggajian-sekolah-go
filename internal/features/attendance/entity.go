package attendance

import (
	"time"
)

const (
	JenisHadir = "hadir"
	JenisKelas = "kelas"
)

func IsValidJenis(j string) bool {
	return j == JenisHadir || j == JenisKelas
}

type AttendanceRecord struct {
	ID            int64      `gorm:"primaryKey"`
	EmployeeID    int64      `gorm:"column:employee_id"`
	Jenis         string     `gorm:"column:jenis"`
	TeachingID    *int64     `gorm:"column:teaching_id"`
	PhotoFilename *string    `gorm:"column:photo_filename"`
	Latitude      float64    `gorm:"column:latitude"`
	Longitude     float64    `gorm:"column:longitude"`
	AccuracyM     *float64   `gorm:"column:accuracy_m"`
	IsMock        bool       `gorm:"column:is_mock"`
	DistanceM     *float64   `gorm:"column:distance_m"`
	DeviceInfo    *string    `gorm:"column:device_info"`
	RecordedAt    time.Time  `gorm:"column:recorded_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (AttendanceRecord) TableName() string { return "attendance_records" }
