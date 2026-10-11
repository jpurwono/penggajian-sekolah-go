// Package config memuat konfigurasi aplikasi dari environment / file .env.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	AppName       string
	AppPort       string
	DatabaseURL   string
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		AppName:       getEnv("APP_NAME", "Penggajian Sekolah API"),
		AppPort:       getEnv("APP_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTAccessTTL:  getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: getDuration("JWT_REFRESH_TTL", 168*time.Hour),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL wajib diisi (lihat .env.example)")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi (lihat .env.example)")
	}
	if cfg.AppEnv == "production" && len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET minimal 32 karakter di production")
	}

	return cfg, nil
}

func (c *Config) IsDev() bool { return c.AppEnv == "development" }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
