package auth

import (
	"errors"

	"github.com/labstack/echo/v4"

	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func mapAuthError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return httpx.ErrUnauthorized(c, err.Error())
	case errors.Is(err, ErrSessionNotFound):
		return httpx.ErrUnauthorized(c, "Sesi tidak valid, silakan masuk kembali")
	case errors.Is(err, ErrWrongPassword), errors.Is(err, ErrUserNotFound):
		return httpx.ErrBadRequest(c, err.Error())
	default:
		return httpx.ErrInternal(c, "Terjadi kesalahan, coba lagi")
	}
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}

	result, err := h.svc.Login(
		c.Request().Context(), req.Username, req.Password,
		c.Request().UserAgent(), c.RealIP(),
	)
	if err != nil {
		return mapAuthError(c, err)
	}
	return httpx.OKWithMessage(c, "Login berhasil", result)
}

func (h *Handler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}

	result, err := h.svc.Refresh(
		c.Request().Context(), req.RefreshToken,
		c.Request().UserAgent(), c.RealIP(),
	)
	if err != nil {
		return mapAuthError(c, err)
	}
	return httpx.OK(c, result)
}

func (h *Handler) Logout(c echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}

	if err := h.svc.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return mapAuthError(c, err)
	}
	return httpx.OKWithMessage(c, "Berhasil keluar", nil)
}

func (h *Handler) Me(c echo.Context) error {
	user, err := h.svc.Me(c.Request().Context(), middleware.CurrentUserID(c))
	if err != nil {
		return mapAuthError(c, err)
	}
	return httpx.OK(c, user)
}

func (h *Handler) ChangePassword(c echo.Context) error {
	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}

	err := h.svc.ChangePassword(c.Request().Context(), middleware.CurrentUserID(c), req.CurrentPassword, req.NewPassword)
	if err != nil {
		return mapAuthError(c, err)
	}
	return httpx.OKWithMessage(c, "Password berhasil diubah, silakan login kembali", nil)
}
