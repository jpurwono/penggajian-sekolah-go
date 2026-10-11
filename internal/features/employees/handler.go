package employees

import (
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return httpx.ErrNotFound(c, ErrNotFound.Error())
	default:
		return httpx.ErrBadRequest(c, err.Error())
	}
}

func (h *Handler) List(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)

	filter := ListFilter{Page: page, PerPage: per, Q: q}
	if v := c.QueryParam("employee_type"); v != "" {
		filter.EmployeeType = v
	}
	if v := c.QueryParam("is_active"); v == "true" || v == "false" {
		b := v == "true"
		filter.Active = &b
	}

	result, err := h.svc.List(c.Request().Context(), filter)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Employees, result.Meta)
}

func (h *Handler) Get(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) Create(c echo.Context) error {
	var req SaveEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.Create(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, resp)
}

func (h *Handler) Update(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.Update(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) Delete(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Pegawai berhasil dihapus", nil)
}
