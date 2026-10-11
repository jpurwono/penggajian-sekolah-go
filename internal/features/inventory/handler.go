package inventory

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"

	"penggajian/server/internal/platform/httpx"
)

type Handler struct {
	master *MasterService
}

func NewHandler(master *MasterService) *Handler {
	return &Handler{master: master}
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return httpx.ErrNotFound(c, ErrNotFound.Error())
	default:
		return httpx.ErrBadRequest(c, err.Error())
	}
}

func queryInt64(c echo.Context, key string) *int64 {
	v := c.QueryParam(key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return nil
	}
	return &n
}

// ---------- Kategori ----------

func (h *Handler) ListCategories(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	data, total, err := h.master.ListCategories(c.Request().Context(), q, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, data, httpx.NewMeta(page, per, total))
}

func (h *Handler) CreateCategory(c echo.Context) error {
	var req SaveCategoryRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.master.CreateCategory(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateCategory(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveCategoryRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	if err := h.master.UpdateCategory(c.Request().Context(), id, req); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Kategori berhasil diperbarui", nil)
}

func (h *Handler) DeleteCategory(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.master.DeleteCategory(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Kategori berhasil dihapus", nil)
}

// ---------- Lokasi ----------

func (h *Handler) ListLocations(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	data, total, err := h.master.ListLocations(c.Request().Context(), q, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, data, httpx.NewMeta(page, per, total))
}

func (h *Handler) CreateLocation(c echo.Context) error {
	var req SaveLocationRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.master.CreateLocation(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateLocation(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveLocationRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.master.UpdateLocation(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteLocation(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.master.DeleteLocation(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Lokasi berhasil dihapus", nil)
}

// ---------- Item ----------

func (h *Handler) ListItems(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)

	var active *bool
	if v := c.QueryParam("is_active"); v == "true" || v == "false" {
		b := v == "true"
		active = &b
	}

	data, total, err := h.master.ListItems(c.Request().Context(), q, queryInt64(c, "category_id"), active, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, data, httpx.NewMeta(page, per, total))
}

func (h *Handler) CreateItem(c echo.Context) error {
	var req SaveItemRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.master.CreateItem(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateItem(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveItemRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.master.UpdateItem(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteItem(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.master.DeleteItem(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMessage(c, "Barang berhasil dihapus", nil)
}
