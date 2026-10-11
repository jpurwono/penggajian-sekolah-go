// Package httpx menyediakan envelope respons standar & error handler Echo.
package httpx

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Meta digunakan untuk metadata paginasi pada respons list.
type Meta struct {
	Page      int64  `json:"page"`
	PerPage   int64  `json:"per_page"`
	Total     int64  `json:"total"`
	TotalPage int64  `json:"total_page"`
}

type envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, envelope{Success: true, Data: data})
}

func OKWithMessage(c echo.Context, message string, data any) error {
	return c.JSON(http.StatusOK, envelope{Success: true, Message: message, Data: data})
}

func OKWithMeta(c echo.Context, data any, meta any) error {
	return c.JSON(http.StatusOK, envelope{Success: true, Data: data, Meta: meta})
}

func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, envelope{Success: true, Data: data})
}

func ErrBadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, envelope{Success: false, Message: message})
}

func ErrUnauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, envelope{Success: false, Message: message})
}

func ErrForbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, envelope{Success: false, Message: message})
}

func ErrNotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, envelope{Success: false, Message: message})
}

func ErrInternal(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, envelope{Success: false, Message: message})
}

// ErrorHandler memastikan semua error (termasuk milik Echo) memakai envelope standar.
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		msg, _ := httpErr.Message.(string)
		if msg == "" {
			msg = http.StatusText(httpErr.Code)
		}
		_ = c.JSON(httpErr.Code, envelope{Success: false, Message: msg})
		return
	}

	_ = c.JSON(http.StatusInternalServerError, envelope{
		Success: false,
		Message: "Terjadi kesalahan internal server",
	})
}
