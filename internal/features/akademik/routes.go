package akademik

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	mw "penggajian/server/internal/platform/middleware"
)

// Register mendaftarkan routes fitur akademik ke group /api/v1.
func Register(g *echo.Group, db *gorm.DB, authMW echo.MiddlewareFunc) {
	repo := NewRepository(db)
	h := NewHandler(NewService(repo), db)

	subjects := g.Group("/akademik/subjects", authMW, mw.RequirePermission("akademik.read"))
	subjects.GET("", h.ListSubjects)
	subjects.POST("", h.CreateSubject, mw.RequirePermission("akademik.create"))
	subjects.PUT("/:id", h.UpdateSubject, mw.RequirePermission("akademik.update"))
	subjects.DELETE("/:id", h.DeleteSubject, mw.RequirePermission("akademik.delete"))

	classes := g.Group("/akademik/classes", authMW, mw.RequirePermission("akademik.read"))
	classes.GET("", h.ListClasses)
	classes.POST("", h.CreateClass, mw.RequirePermission("akademik.create"))
	classes.PUT("/:id", h.UpdateClass, mw.RequirePermission("akademik.update"))
	classes.DELETE("/:id", h.DeleteClass, mw.RequirePermission("akademik.delete"))

	teachings := g.Group("/akademik/teachings", authMW, mw.RequirePermission("pegawai.read"))
	teachings.PUT("/:id", h.UpdateTeaching, mw.RequirePermission("pegawai.update"))
	teachings.DELETE("/:id", h.DeleteTeaching, mw.RequirePermission("pegawai.delete"))

	empTeachings := g.Group("/employees", authMW, mw.RequirePermission("pegawai.read"))
	empTeachings.GET("/:id/teachings", h.ListTeachings)
	empTeachings.POST("/:id/teachings", h.CreateTeaching, mw.RequirePermission("pegawai.create"))

	// Khusus aplikasi guru: jadwal milik sendiri (cukup login).
	g.GET("/akademik/my-teachings", h.MyTeachings, authMW)
}
