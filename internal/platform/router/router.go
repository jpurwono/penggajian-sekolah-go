// Package router merakit Echo dan mendaftarkan routes seluruh fitur.
// Fitur baru cukup ditambahkan di sini — satu blok per fitur.
package router

import (
	"context"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"

	"penggajian/api/internal/features/akademik"
	"penggajian/api/internal/features/attendance"
	"penggajian/api/internal/features/auth"
	"penggajian/api/internal/features/employees"
	"penggajian/api/internal/features/health"
	"penggajian/api/internal/features/inventory"
	"penggajian/api/internal/features/payroll"
	"penggajian/api/internal/features/rbac"
	"penggajian/api/internal/features/reports"
	"penggajian/api/internal/features/settings"
	"penggajian/api/internal/features/spp"
	"penggajian/api/internal/platform/config"
	"penggajian/api/internal/platform/httpx"
	"penggajian/api/internal/platform/middleware"
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
	payrollRepo := payroll.NewRepository(db)
	payrollSvc := payroll.NewService(payrollRepo)

	// Pegawai baru otomatis menerima komponen dari template default sesuai tipe pegawainya.
	employees.SetPostCreateHook(func(ctx context.Context, emp *employees.Employee) error {
		return payrollSvc.ApplyDefaultTemplate(ctx, emp.ID, emp.EmployeeType)
	})

	employees.Register(v1, employees.NewService(employees.NewRepository(db)), authMW)

	payroll.Register(
		v1,
		payrollSvc,
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
