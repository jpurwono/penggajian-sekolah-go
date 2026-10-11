package middleware

import (
	"context"
	"slices"
	"strings"

	"github.com/labstack/echo/v4"

	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/security"
)

const contextUserKey = "auth_user"

// AuthUser adalah identitas user yang menempel di context request.
type AuthUser struct {
	ID          int64
	Username    string
	Email       string
	FullName    string
	RoleID      int64
	RoleName    string
	EmployeeID  *int64
	Permissions []string
}

// UserLoader memuat user aktif berdasarkan ID.
// Diimplementasikan oleh repository fitur auth agar middleware tidak bergantung ke fitur tertentu.
type UserLoader interface {
	LoadActiveUserByID(ctx context.Context, id int64) (*AuthUser, error)
}

// Auth memverifikasi Bearer token dan menyuntikkan user (beserta permission) ke context.
// Token bisa dari header Authorization ATAU query param ?token= (untuk <img> tag).
func Auth(loader UserLoader, jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenStr := ""

			header := c.Request().Header.Get(echo.HeaderAuthorization)
			const prefix = "Bearer "
			if strings.HasPrefix(header, prefix) {
				tokenStr = strings.TrimPrefix(header, prefix)
			} else if qToken := c.QueryParam("token"); qToken != "" {
				tokenStr = qToken
			}

			if tokenStr == "" {
				return httpx.ErrUnauthorized(c, "Token akses tidak ditemukan")
			}

			userID, err := security.ParseAccessToken(jwtSecret, tokenStr)
			if err != nil {
				return httpx.ErrUnauthorized(c, "Sesi berakhir, silakan masuk kembali")
			}

			user, err := loader.LoadActiveUserByID(c.Request().Context(), userID)
			if err != nil {
				return httpx.ErrUnauthorized(c, "User tidak aktif atau tidak ditemukan")
			}

			c.Set(contextUserKey, user)
			return next(c)
		}
	}
}

// RequirePermission menolak akses jika user tidak punya permission tertentu.
func RequirePermission(code string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := CurrentUser(c)
			if user == nil {
				return httpx.ErrUnauthorized(c, "Belum terautentikasi")
			}
			if !slices.Contains(user.Permissions, code) {
				return httpx.ErrForbidden(c, "Anda tidak memiliki izin untuk aksi ini")
			}
			return next(c)
		}
	}
}

// RequireAnyPermission mengizinkan akses jika user punya salah satu dari permission tersebut.
func RequireAnyPermission(codes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := CurrentUser(c)
			if user == nil {
				return httpx.ErrUnauthorized(c, "Belum terautentikasi")
			}
			for _, code := range codes {
				if slices.Contains(user.Permissions, code) {
					return next(c)
				}
			}
			return httpx.ErrForbidden(c, "Anda tidak memiliki izin untuk aksi ini")
		}
	}
}

func CurrentUser(c echo.Context) *AuthUser {
	user, _ := c.Get(contextUserKey).(*AuthUser)
	return user
}

func CurrentUserID(c echo.Context) int64 {
	if user := CurrentUser(c); user != nil {
		return user.ID
	}
	return 0
}
