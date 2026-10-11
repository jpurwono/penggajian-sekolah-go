package settings

import (
	"time"
)

type SchoolProfile struct {
	ID            int64     `gorm:"primaryKey"`
	SchoolName    string    `gorm:"column:school_name"`
	Address       *string   `gorm:"column:address"`
	Phone         *string   `gorm:"column:phone"`
	Email         *string   `gorm:"column:email"`
	PrincipalName *string   `gorm:"column:principal_name"`
	LogoURL       *string   `gorm:"column:logo_url"`
	WorkingDays   int       `gorm:"column:working_days"`
	Latitude      *float64  `gorm:"column:latitude"`
	Longitude     *float64  `gorm:"column:longitude"`
	GeofenceRadiusM int     `gorm:"column:geofence_radius_m"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (SchoolProfile) TableName() string { return "school_profiles" }
