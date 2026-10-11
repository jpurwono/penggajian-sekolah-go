// Package middleware berisi middleware HTTP lintas fitur.
package middleware

import (
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// CORS mengizinkan akses dari origin web yang terdaftar.
// Origin tambahan bisa diset via environment CORS_ORIGINS (dipisah koma).
func CORS() echo.MiddlewareFunc {
	origins := []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://100.86.43.118:3002",
		"http://100.86.43.118:8090",
		"https://smpmawaddah.levobyte.com",
	}
	if env := os.Getenv("CORS_ORIGINS"); env != "" {
		origins = append(origins, strings.Split(env, ",")...)
	}
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	})
}
