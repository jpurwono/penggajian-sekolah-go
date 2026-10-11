package spp

import (
	"time"

	"gorm.io/gorm"
)

const (
	FreqMonthly = "monthly"
	FreqAnnual  = "annual"
)

func IsValidFrequency(f string) bool { return f == FreqMonthly || f == FreqAnnual }

type FeeType struct {
	ID             int64          `gorm:"primaryKey"`
	Code           string         `gorm:"column:code"`
	Name           string         `gorm:"column:name"`
	Frequency      string         `gorm:"column:frequency"`
	DefaultAmount  int64          `gorm:"column:default_amount"`
	Description    *string        `gorm:"column:description"`
	IsActive       bool           `gorm:"column:is_active"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (FeeType) TableName() string { return "spp_fee_types" }
