package rbac

import (
	"github.com/labstack/echo/v4"

	mw "penggajian/server/internal/platform/middleware"
)

// Register mendaftarkan routes fitur RBAC ke group /api/v1.
func Register(g *echo.Group, svc *Service, authMW echo.MiddlewareFunc) {
	h := NewHandler(svc)

	users := g.Group("/users", authMW, mw.RequirePermission("pengaturan.read"))
	users.GET("", h.ListUsers)
	users.GET("/:id", h.GetUser)
	users.POST("", h.CreateUser, mw.RequirePermission("pengaturan.create"))
	users.PUT("/:id", h.UpdateUser, mw.RequirePermission("pengaturan.update"))
	users.PUT("/:id/password", h.UpdateUserPassword, mw.RequirePermission("pengaturan.update"))
	users.DELETE("/:id", h.DeleteUser, mw.RequirePermission("pengaturan.delete"))

	roles := g.Group("/roles", authMW, mw.RequirePermission("pengaturan.read"))
	roles.GET("", h.ListRoles)
	roles.GET("/:id", h.GetRole)
	roles.POST("", h.CreateRole, mw.RequirePermission("pengaturan.create"))
	roles.PUT("/:id", h.UpdateRole, mw.RequirePermission("pengaturan.update"))
	roles.DELETE("/:id", h.DeleteRole, mw.RequirePermission("pengaturan.delete"))

	g.GET("/permissions", h.ListPermissions, authMW, mw.RequirePermission("pengaturan.read"))
}
