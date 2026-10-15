package payroll

import (
	"github.com/labstack/echo/v4"

	mw "penggajian/server/internal/platform/middleware"
)

// Register mendaftarkan routes fitur payroll ke group /api/v1.
func Register(g *echo.Group, svc *Service, slipSvc *SlipService, authMW echo.MiddlewareFunc) {
	h := NewHandler(svc)
	sh := NewSlipHandler(slipSvc)

	components := g.Group("/payroll/components", authMW, mw.RequirePermission("penggajian.read"))
	components.GET("", h.ListComponents)
	components.GET("/:id", h.GetComponent)
	components.POST("", h.CreateComponent, mw.RequirePermission("penggajian.create"))
	components.PUT("/:id", h.UpdateComponent, mw.RequirePermission("penggajian.update"))
	components.DELETE("/:id", h.DeleteComponent, mw.RequirePermission("penggajian.delete"))

	empComponents := g.Group("/payroll/employee-components", authMW, mw.RequirePermission("penggajian.read"))
	empComponents.GET("", h.ListEmployeeComponents)
	empComponents.POST("", h.AddEmployeeComponent, mw.RequirePermission("penggajian.create"))
	empComponents.PUT("/:id", h.UpdateEmployeeComponent, mw.RequirePermission("penggajian.update"))
	empComponents.DELETE("/:id", h.DeleteEmployeeComponent, mw.RequirePermission("penggajian.delete"))

	periods := g.Group("/payroll/periods", authMW, mw.RequirePermission("penggajian.read"))
	periods.GET("", h.ListPeriods)
	periods.GET("/:id", h.GetPeriod)
	periods.POST("", h.CreatePeriod, mw.RequirePermission("penggajian.create"))
	periods.PUT("/:id", h.UpdatePeriod, mw.RequirePermission("penggajian.update"))
	periods.DELETE("/:id", h.DeletePeriod, mw.RequirePermission("penggajian.delete"))
	periods.POST("/:id/generate", sh.Generate, mw.RequirePermission("penggajian.create"))
	periods.PUT("/:id/attendance", sh.UpdateAttendance, mw.RequirePermission("penggajian.update"))
	periods.POST("/:id/attendance/import", sh.ImportAttendance, mw.RequirePermission("penggajian.update"))
	periods.POST("/:id/approve", h.ApprovePeriod, mw.RequirePermission("penggajian.approve"))
	periods.POST("/:id/pay", h.PayPeriod, mw.RequirePermission("penggajian.approve"))
	periods.POST("/:id/cancel", h.CancelPeriod, mw.RequirePermission("penggajian.approve"))
	periods.GET("/:id/payslips", sh.ListSlipsByPeriod)
	periods.GET("/:id/recap", sh.Recap)

	payslips := g.Group("/payroll/payslips", authMW, mw.RequirePermission("penggajian.read"))
	payslips.GET("/:id", sh.GetSlip)
	payslips.PUT("/:id", sh.UpdateSlipNote, mw.RequirePermission("penggajian.update"))
	payslips.POST("/:id/items", sh.AddItem, mw.RequirePermission("penggajian.update"))
	payslips.PUT("/:id/items/:item_id", sh.UpdateItem, mw.RequirePermission("penggajian.update"))
	payslips.DELETE("/:id/items/:item_id", sh.DeleteItem, mw.RequirePermission("penggajian.update"))
}
