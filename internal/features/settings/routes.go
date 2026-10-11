package settings

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/middleware"
)

// Register mendaftarkan routes fitur settings ke group /api/v1.
func Register(g *echo.Group, db *gorm.DB, authMW echo.MiddlewareFunc) {
	h := NewHandler(db)

	grp := g.Group("/settings", authMW)
	grp.GET("/school-profile", h.Get)
	grp.PUT("/school-profile", h.Update, middleware.RequirePermission("pengaturan.update"))
}
