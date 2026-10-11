package spp

import (
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
)

type FeeTypeHandler struct {
	svc *FeeTypeService
}

func NewFeeTypeHandler(svc *FeeTypeService) *FeeTypeHandler {
	return &FeeTypeHandler{svc: svc}
}

func (h *FeeTypeHandler) List(c echo.Context) error {
	var active *bool
	if v := c.QueryParam("is_active"); v == "true" || v == "false" {
		b := v == "true"
		active = &b
	}
	data, err := h.svc.List(c.Request().Context(), c.QueryParam("q"), active)
	if err != nil {
		return c.JSON(500, echo.Map{"success": false, "message": "Gagal memuat jenis biaya"})
	}
	return httpx.OK(c, data)
}

func (h *FeeTypeHandler) Create(c echo.Context) error {
	var req SaveFeeTypeRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.Create(c.Request().Context(), c, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpx.ErrNotFound(c, "Jenis biaya tidak ditemukan")
		}
		return httpx.ErrBadRequest(c, err.Error())
	}
	return httpx.Created(c, resp)
}

func (h *FeeTypeHandler) Update(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveFeeTypeRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.Update(c.Request().Context(), c, id, req)
	if err != nil {
		return httpx.ErrBadRequest(c, err.Error())
	}
	return httpx.OK(c, resp)
}

func (h *FeeTypeHandler) Delete(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.Delete(c.Request().Context(), c, id); err != nil {
		return httpx.ErrBadRequest(c, err.Error())
	}
	return httpx.OKWithMessage(c, "Jenis biaya berhasil dihapus", nil)
}

func (h *FeeTypeHandler) GenerateYearly(c echo.Context) error {
	var req GenerateYearlyRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	result, err := h.svc.GenerateYearly(c.Request().Context(), c, req)
	if err != nil {
		return httpx.ErrBadRequest(c, err.Error())
	}
	return httpx.OKWithMessage(c, result.Message, result)
}
