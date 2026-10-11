package employees

import (
	"github.com/labstack/echo/v4"

	mw "penggajian/server/internal/platform/middleware"
)

// Register mendaftarkan routes fitur employees ke group /api/v1.
func Register(g *echo.Group, svc *Service, authMW echo.MiddlewareFunc) {
	h := NewHandler(svc)

	grp := g.Group("/employees", authMW, mw.RequirePermission("pegawai.read"))
	grp.GET("", h.List)
	grp.GET("/:id", h.Get)
	grp.POST("", h.Create, mw.RequirePermission("pegawai.create"))
	grp.PUT("/:id", h.Update, mw.RequirePermission("pegawai.update"))
	grp.DELETE("/:id", h.Delete, mw.RequirePermission("pegawai.delete"))
}
