package payroll

import (
	"github.com/labstack/echo/v4"

	"penggajian/api/internal/platform/audit"
	"penggajian/api/internal/platform/httpx"
)

func (h *Handler) ListTemplates(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	result, err := h.svc.ListTemplates(c.Request().Context(), page, per, q)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Templates, result.Meta)
}

func (h *Handler) GetTemplate(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.GetTemplate(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) CreateTemplate(c echo.Context) error {
	var req SaveTemplateRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreateTemplate(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "create", "payroll_templates", resp.ID, nil, resp)
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateTemplate(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	old, _ := h.svc.GetTemplate(c.Request().Context(), id)

	var req SaveTemplateRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateTemplate(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "update", "payroll_templates", id, old, resp)
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteTemplate(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	old, _ := h.svc.GetTemplate(c.Request().Context(), id)

	if err := h.svc.DeleteTemplate(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "delete", "payroll_templates", id, old, nil)
	return httpx.OKWithMessage(c, "Template berhasil dihapus", nil)
}

func (h *Handler) ApplyTemplate(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req ApplyTemplateRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.ApplyTemplate(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "update", "employee_components", id,
		map[string]any{"template_id": id, "request": req}, resp)
	return httpx.OK(c, resp)
}
