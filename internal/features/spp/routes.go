package spp

import (
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	mw "penggajian/server/internal/platform/middleware"
)

func timeNowYear() int { return time.Now().Year() }

// Register mendaftarkan routes fitur SPP ke group /api/v1.
func Register(g *echo.Group, db *gorm.DB, authMW echo.MiddlewareFunc) {
	repo := NewRepository(db)
	h := NewHandler(NewService(db, repo))

	students := g.Group("/spp/students", authMW, mw.RequirePermission("spp.read"))
	students.GET("", h.ListStudents)
	students.POST("", h.CreateStudent, mw.RequirePermission("spp.create"))
	students.PUT("/:id", h.UpdateStudent, mw.RequirePermission("spp.update"))
	students.DELETE("/:id", h.DeleteStudent, mw.RequirePermission("spp.delete"))

	bills := g.Group("/spp/bills", authMW, mw.RequirePermission("spp.read"))
	bills.GET("", h.ListBills)
	bills.GET("/:id", h.GetBill)
	bills.POST("", h.CreateBill, mw.RequirePermission("spp.create"))
	bills.POST("/generate", h.GenerateBills, mw.RequirePermission("spp.create"))
	bills.PUT("/:id", h.UpdateBill, mw.RequirePermission("spp.update"))
	bills.DELETE("/:id", h.DeleteBill, mw.RequirePermission("spp.delete"))
	bills.GET("/:id/payments", h.ListBillPayments)
	bills.POST("/:id/payments", h.CreatePayment, mw.RequirePermission("spp.update"))

	payments := g.Group("/spp/payments", authMW, mw.RequirePermission("spp.read"))
	payments.GET("", h.ListPayments)
	payments.GET("/:id", h.GetPayment)

	feeTypes := g.Group("/spp/fee-types", authMW, mw.RequirePermission("spp.read"))
	feeTypeSvc := NewFeeTypeService(db, NewFeeTypeRepository(db), NewRepository(db))
	fth := NewFeeTypeHandler(feeTypeSvc)
	feeTypes.GET("", fth.List)
	feeTypes.POST("", fth.Create, mw.RequirePermission("spp.create"))
	feeTypes.PUT("/:id", fth.Update, mw.RequirePermission("spp.update"))
	feeTypes.DELETE("/:id", fth.Delete, mw.RequirePermission("spp.delete"))
	g.POST("/spp/bills/generate-yearly", fth.GenerateYearly, authMW, mw.RequirePermission("spp.create"))

	g.GET("/spp/summary", h.YearSummary, authMW, mw.RequirePermission("spp.read"))
}
