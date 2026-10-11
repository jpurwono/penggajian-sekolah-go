// Package health adalah contoh pola fitur: handler + routes, mandiri.
package health

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/config"
)

type Handler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{db: db, cfg: cfg}
}

func (h *Handler) Check(c echo.Context) error {
	dbStatus := "up"
	if sqlDB, err := h.db.DB(); err != nil {
		dbStatus = "down"
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "down"
	}

	status := http.StatusOK
	if dbStatus == "down" {
		status = http.StatusServiceUnavailable
	}

	return c.JSON(status, echo.Map{
		"success": dbStatus == "up",
		"data": echo.Map{
			"app":      h.cfg.AppName,
			"env":      h.cfg.AppEnv,
			"database": dbStatus,
			"time":     time.Now().Format(time.RFC3339),
		},
	})
}
