package rbac

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          int64 `gorm:"primaryKey"`
	Name        string
	Description string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type Permission struct {
	ID        int64 `gorm:"primaryKey"`
	Code      string
	Name      string
	Module    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID           int64 `gorm:"primaryKey"`
	Username     string
	Email        string
	PasswordHash string
	FullName     string
	RoleID       int64
	Role         *Role `gorm:"foreignKey:RoleID"`
	EmployeeID   *int64
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
