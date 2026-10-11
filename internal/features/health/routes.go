package health

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/config"
)

// Register mendaftarkan routes fitur health ke group /api/v1.
func Register(g *echo.Group, db *gorm.DB, cfg *config.Config) {
	h := NewHandler(db, cfg)
	g.GET("/health", h.Check)
}
