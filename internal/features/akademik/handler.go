package akademik

import (
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/middleware"
)

type Handler struct {
	svc *Service
	db  *gorm.DB
}

func NewHandler(svc *Service, db *gorm.DB) *Handler {
	return &Handler{svc: svc, db: db}
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return httpx.ErrNotFound(c, ErrNotFound.Error())
	default:
		return httpx.ErrBadRequest(c, err.Error())
	}
}

// ---------- Mapel ----------

func (h *Handler) ListSubjects(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	result, err := h.svc.ListSubjects(c.Request().Context(), page, per, q)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Subjects, result.Meta)
}

func (h *Handler) CreateSubject(c echo.Context) error {
	var req SaveSubjectRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreateSubject(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "create", "subjects", resp.ID, nil, resp)
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateSubject(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveSubjectRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateSubject(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "update", "subjects", id, nil, resp)
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteSubject(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.DeleteSubject(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "delete", "subjects", id, nil, nil)
	return httpx.OKWithMessage(c, "Mapel berhasil dihapus", nil)
}

// ---------- Kelas ----------

func (h *Handler) ListClasses(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	result, err := h.svc.ListClasses(c.Request().Context(), page, per, q)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, result.Classes, result.Meta)
}

func (h *Handler) CreateClass(c echo.Context) error {
	var req SaveClassRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreateClass(c.Request().Context(), req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "create", "school_classes", resp.ID, nil, resp)
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateClass(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveClassRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateClass(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "update", "school_classes", id, nil, resp)
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteClass(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.DeleteClass(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "delete", "school_classes", id, nil, nil)
	return httpx.OKWithMessage(c, "Kelas berhasil dihapus", nil)
}

// ---------- Penugasan Mengajar ----------

func (h *Handler) ListTeachings(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID pegawai tidak valid")
	}
	rows, err := h.svc.ListTeachings(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, rows)
}

func (h *Handler) CreateTeaching(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID pegawai tidak valid")
	}
	var req SaveTeachingRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.CreateTeaching(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "create", "teaching_assignments", resp.ID, nil, resp)
	return httpx.Created(c, resp)
}

func (h *Handler) UpdateTeaching(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveTeachingRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.UpdateTeaching(c.Request().Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "update", "teaching_assignments", id, nil, resp)
	return httpx.OK(c, resp)
}

func (h *Handler) DeleteTeaching(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.DeleteTeaching(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	audit.Log(c, h.db, "delete", "teaching_assignments", id, nil, nil)
	return httpx.OKWithMessage(c, "Penugasan mengajar berhasil dihapus", nil)
}

// MyTeachings: jadwal mengajar milik pegawai yang terhubung dengan user login.
// Tanpa permission khusus — cukup autentikasi (dipakai aplikasi Absen Guru).
func (h *Handler) MyTeachings(c echo.Context) error {
	user := middleware.CurrentUser(c)
	if user == nil || user.EmployeeID == nil || *user.EmployeeID <= 0 {
		return httpx.OK(c, echo.Map{"employee_id": nil, "teachings": []any{}})
	}
	rows, err := h.svc.ListTeachings(c.Request().Context(), *user.EmployeeID)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, echo.Map{"employee_id": *user.EmployeeID, "teachings": rows})
}
