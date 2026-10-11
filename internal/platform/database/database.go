// Package database menyediakan koneksi PostgreSQL, runner migrasi SQL, dan seed awal.
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	logLevel := logger.Warn
	if os.Getenv("APP_ENV") == "" || os.Getenv("APP_ENV") == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger:  logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("koneksi database gagal: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database gagal: %w", err)
	}

	log.Println("[db] tersambung ke PostgreSQL")
	return db, nil
}
