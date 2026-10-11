package auth

import (
	"github.com/labstack/echo/v4"
)

// Register mendaftarkan routes fitur auth ke group /api/v1.
func Register(g *echo.Group, svc *Service, authMW echo.MiddlewareFunc) {
	h := NewHandler(svc)

	grp := g.Group("/auth")
	grp.POST("/login", h.Login)
	grp.POST("/refresh", h.Refresh)

	authed := grp.Group("", authMW)
	authed.GET("/me", h.Me)
	authed.POST("/logout", h.Logout)
	authed.PUT("/password", h.ChangePassword)
}
