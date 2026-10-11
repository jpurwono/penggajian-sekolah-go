package attendance

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

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

// POST /attendance — guru absen (dari aplikasi Android, multipart form: JSON + foto opsional)
func (h *Handler) CheckIn(c echo.Context) error {
	var req CheckInRequest

	// dukung JSON body atau multipart form
	contentType := c.Request().Header.Get(echo.HeaderContentType)
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req.Jenis = c.FormValue("jenis")
		req.Latitude = parseF64(c.FormValue("latitude"))
		req.Longitude = parseF64(c.FormValue("longitude"))
		req.AccuracyM = parseF64Ptr(c.FormValue("accuracy_m"))
		req.IsMock = c.FormValue("is_mock") == "true"
		req.RecordedAt = c.FormValue("recorded_at")
		req.DeviceInfo = c.FormValue("device_info")
		if v := c.FormValue("teaching_id"); v != "" {
			id := parseI64(v)
			req.TeachingID = &id
		}
	} else {
		if err := c.Bind(&req); err != nil {
			return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
		}
	}

	result, err := h.svc.CheckIn(c.Request().Context(), c, req)
	if err != nil {
		return mapError(c, err)
	}

	// upload foto jika ada (multipart)
	if file, err := c.FormFile("photo"); err == nil {
		src, err := file.Open()
		if err != nil {
			return httpx.ErrBadRequest(c, "Gagal membaca foto")
		}
		defer src.Close()

		dir := "uploads/attendance"
		os.MkdirAll(dir, 0o755)
		filename := "att_" + strconv.FormatInt(result.Record.ID, 10) + ".jpg"
		dst, err := os.Create(filepath.Join(dir, filename))
		if err != nil {
			return httpx.ErrInternal(c, "Gagal menyimpan foto")
		}
		defer dst.Close()
		io.Copy(dst, src)

		if err := h.svc.repo.SavePhoto(c.Request().Context(), result.Record.ID, filename); err != nil {
			return httpx.ErrInternal(c, "Gagal menyimpan referensi foto")
		}
		result.Record.HasPhoto = true
	}

	return httpx.Created(c, result)
}

// GET /attendance — list (admin web)
func (h *Handler) List(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	_ = q

	f := Filter{
		Jenis:    c.QueryParam("jenis"),
		DateFrom: c.QueryParam("date_from"),
		DateTo:   c.QueryParam("date_to"),
	}
	if v := c.QueryParam("employee_id"); v != "" {
		id := parseI64(v)
		if id > 0 {
			f.EmployeeID = &id
		}
	}

	rows, meta, err := h.svc.List(c.Request().Context(), f, page, per)
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OKWithMeta(c, rows, meta)
}

// GET /attendance/:id — detail
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

// GET /attendance/daily-summary — rekap harian semua pegawai
func (h *Handler) DailySummary(c echo.Context) error {
	resp, err := h.svc.DailySummary(c.Request().Context(), c.QueryParam("date"))
	if err != nil {
		return mapError(c, err)
	}
	return httpx.OK(c, resp)
}

// GET /attendance/photo/:id — serve foto
func (h *Handler) Photo(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	filename, err := h.svc.repo.PhotoPath(c.Request().Context(), id)
	if err != nil || filename == "" {
		return httpx.ErrNotFound(c, "Foto tidak ditemukan")
	}
	path := filepath.Join("uploads/attendance", filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return httpx.ErrNotFound(c, "File foto tidak ditemukan")
	}
	return c.File(path)
}

// Register mendaftarkan routes fitur attendance ke group /api/v1.
func Register(g *echo.Group, db *gorm.DB, authMW echo.MiddlewareFunc) {
	repo := NewRepository(db)
	h := NewHandler(NewService(db, repo), db)

	// Guru: absen sendiri (cukup login, tanpa permission khusus)
	g.POST("/attendance", h.CheckIn, authMW)

	// Admin: monitoring
	att := g.Group("/attendance", authMW, middleware.RequirePermission("pegawai.read"))
	att.GET("", h.List)
	att.GET("/daily-summary", h.DailySummary)
	att.GET("/photo/:id", h.Photo)
	att.GET("/:id", h.Get)
}

func parseF64(v string) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func parseF64Ptr(v string) *float64 {
	if v == "" {
		return nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &f
}

func parseI64(v string) int64 {
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}
