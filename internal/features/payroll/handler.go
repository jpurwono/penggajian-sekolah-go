package payroll

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/middleware"
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

func (h *Handler) ListComponents(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	f := ComponentFilter{Page: page, PerPage: per, Q: q, Type: c.QueryParam("type")}
	if v := c.QueryParam("is_active"); v == "true" || v == "false" {
		b := v == "true"
		f.Active = &b
	}
	result, err := h.svc.ListComponents(c.Request().Context(), f)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Components, result.Meta)
}

func (h *Handler) GetComponent(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.GetComponent(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) CreateComponent(c echo.Context) error {
	var req SaveComponentRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreateComponent(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "create", "pay_components", resp.ID, nil, resp)
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateComponent(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	old, _ := h.svc.GetComponent(c.Request().Context(), id)

	var req SaveComponentRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateComponent(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "update", "pay_components", id, old, resp)
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteComponent(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	old, _ := h.svc.GetComponent(c.Request().Context(), id)

	if err := h.svc.DeleteComponent(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "delete", "pay_components", id, old, nil)
	return httpx.OKWithMessage(c, "Komponen berhasil dihapus", nil)
}

func (h *Handler) ListEmployeeComponents(c echo.Context) error {
	page, per, _ := httpx.ParseListParams(c)
	var employeeID int64
	if v := c.QueryParam("employee_id"); v != "" {
		employeeID = parseInt64(v)
	}
	if employeeID <= 0 {
		return httpx.ErrBadRequest(c, "employee_id wajib diisi")
	}
	result, err := h.svc.ListEmployeeComponents(c.Request().Context(), employeeID, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.EmployeeComponents, result.Meta)
}

func (h *Handler) AddEmployeeComponent(c echo.Context) error {
	var req SaveEmployeeComponentRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.AddEmployeeComponent(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "create", "employee_components", resp.ID, nil, resp)
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateEmployeeComponent(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveEmployeeComponentRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateEmployeeComponent(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "update", "employee_components", id, nil, resp)
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteEmployeeComponent(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.DeleteEmployeeComponent(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "delete", "employee_components", id, nil, nil)
	return httpx.OKWithMessage(c, "Komponen pegawai berhasil dihapus", nil)
}

func (h *Handler) ListPeriods(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	f := PeriodFilter{Page: page, PerPage: per, Q: q, Status: c.QueryParam("status")}
	if v := c.QueryParam("year"); v != "" {
		f.Year = int(parseInt64(v))
	}
	result, err := h.svc.ListPeriods(c.Request().Context(), f)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Periods, result.Meta)
}

func (h *Handler) GetPeriod(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.GetPeriod(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) CreatePeriod(c echo.Context) error {
	var req SavePeriodRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreatePeriod(c.Request().Context(), req, middleware.CurrentUserID(c))
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "create", "payroll_periods", resp.ID, nil, resp)
	return httpx.Created(c, resp)
}

func (h *Handler) UpdatePeriod(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SavePeriodRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdatePeriod(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "update", "payroll_periods", id, nil, resp)
	return httpx.OK(c, resp)
}

func (h *Handler) DeletePeriod(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	old, _ := h.svc.GetPeriod(c.Request().Context(), id)

	if err := h.svc.DeletePeriod(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "delete", "payroll_periods", id, old, nil)
	return httpx.OKWithMessage(c, "Periode berhasil dihapus", nil)
}

func (h *Handler) ApprovePeriod(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.ApprovePeriod(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "approve", "payroll_periods", id, nil, resp)
	return httpx.OKWithMessage(c, "Periode berhasil disetujui", resp)
}

func (h *Handler) PayPeriod(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.PayPeriod(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "pay", "payroll_periods", id, nil, resp)
	return httpx.OKWithMessage(c, "Periode ditandai sudah dibayar", resp)
}

func (h *Handler) CancelPeriod(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	resp, err := h.svc.CancelPeriod(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.svc.repo.db, "cancel", "payroll_periods", id, nil, resp)
	return httpx.OKWithMessage(c, "Periode dibatalkan", resp)
}

func parseInt64(v string) int64 {
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
