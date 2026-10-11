// Package router merakit Echo dan mendaftarkan routes seluruh fitur.
// Fitur baru cukup ditambahkan di sini — satu blok per fitur.
package router

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"

	"penggajian/server/internal/features/akademik"
	"penggajian/server/internal/features/attendance"
	"penggajian/server/internal/features/auth"
	"penggajian/server/internal/features/employees"
	"penggajian/server/internal/features/health"
	"penggajian/server/internal/features/inventory"
	"penggajian/server/internal/features/payroll"
	"penggajian/server/internal/features/rbac"
	"penggajian/server/internal/features/reports"
	"penggajian/server/internal/features/settings"
	"penggajian/server/internal/features/spp"
	"penggajian/server/internal/platform/config"
	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/middleware"
)

func New(cfg *config.Config, db *gorm.DB) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = httpx.ErrorHandler

	e.Use(echomw.Recover())
	if cfg.IsDev() {
		e.Use(echomw.Logger())
	}
	e.Use(middleware.CORS())

	v1 := e.Group("/api/v1")

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(cfg, authRepo)
	authMW := middleware.Auth(authRepo, cfg.JWTSecret)

	health.Register(v1, db, cfg)
	auth.Register(v1, authSvc, authMW)
	rbac.Register(v1, rbac.NewService(rbac.NewRepository(db)), authMW)
	employees.Register(v1, employees.NewService(employees.NewRepository(db)), authMW)

	payrollRepo := payroll.NewRepository(db)
	payroll.Register(
		v1,
		payroll.NewService(payrollRepo),
		payroll.NewSlipService(db, payrollRepo, payroll.NewSlipRepository(db)),
		authMW,
	)
	settings.Register(v1, db, authMW)
	inventory.Register(v1, db, authMW)
	reports.Register(v1, db, authMW)
	spp.Register(v1, db, authMW)
	akademik.Register(v1, db, authMW)
	attendance.Register(v1, db, authMW)

	return e
}
