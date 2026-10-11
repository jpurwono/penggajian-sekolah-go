package settings

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/httpx"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

type SchoolProfileResponse struct {
	SchoolName    string  `json:"school_name"`
	Address       *string `json:"address"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	PrincipalName *string `json:"principal_name"`
	LogoURL       *string `json:"logo_url"`
	WorkingDays   int     `json:"working_days"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	GeofenceRadiusM int   `json:"geofence_radius_m"`
}

type SaveSchoolProfileRequest struct {
	SchoolName    string   `json:"school_name"`
	Address       *string  `json:"address"`
	Phone         *string  `json:"phone"`
	Email         *string  `json:"email"`
	PrincipalName *string  `json:"principal_name"`
	LogoURL       *string  `json:"logo_url"`
	WorkingDays   *int     `json:"working_days"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	GeofenceRadiusM *int   `json:"geofence_radius_m"`
}

func (h *Handler) load(ctx context.Context) (*SchoolProfile, error) {
	var p SchoolProfile
	if err := h.db.WithContext(ctx).First(&p, 1).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func toResponse(p *SchoolProfile) *SchoolProfileResponse {
	return &SchoolProfileResponse{
		SchoolName:    p.SchoolName,
		Address:       p.Address,
		Phone:         p.Phone,
		Email:         p.Email,
		PrincipalName: p.PrincipalName,
		LogoURL:       p.LogoURL,
		WorkingDays:   p.WorkingDays,
		Latitude:      p.Latitude,
		Longitude:     p.Longitude,
		GeofenceRadiusM: p.GeofenceRadiusM,
	}
}

func clean(v *string) *string {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil
	}
	return v
}

func (h *Handler) Get(c echo.Context) error {
	p, err := h.load(c.Request().Context())
	if err != nil {
		return httpx.ErrInternal(c, "Gagal memuat profil sekolah")
	}
	return httpx.OK(c, toResponse(p))
}

func (h *Handler) Update(c echo.Context) error {
	var req SaveSchoolProfileRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	if strings.TrimSpace(req.SchoolName) == "" {
		return httpx.ErrBadRequest(c, "Nama sekolah wajib diisi")
	}
	workingDays := 5
	if req.WorkingDays != nil {
		if *req.WorkingDays != 5 && *req.WorkingDays != 6 {
			return httpx.ErrBadRequest(c, "Hari kerja per minggu harus 5 atau 6")
		}
		workingDays = *req.WorkingDays
	}

	p, err := h.load(c.Request().Context())
	if err != nil {
		return httpx.ErrInternal(c, "Gagal memuat profil sekolah")
	}
	old := toResponse(p)

	p.SchoolName = strings.TrimSpace(req.SchoolName)
	p.Address = clean(req.Address)
	p.Phone = clean(req.Phone)
	p.Email = clean(req.Email)
	p.PrincipalName = clean(req.PrincipalName)
	p.LogoURL = clean(req.LogoURL)
	p.WorkingDays = workingDays
	p.Latitude = req.Latitude
	p.Longitude = req.Longitude
	if req.GeofenceRadiusM != nil && *req.GeofenceRadiusM > 0 {
		p.GeofenceRadiusM = *req.GeofenceRadiusM
	}
	if err := h.db.WithContext(c.Request().Context()).
		Model(p).Select("school_name", "address", "phone", "email", "principal_name", "logo_url", "working_days", "latitude", "longitude", "geofence_radius_m").
		Updates(p).Error; err != nil {
		return httpx.ErrInternal(c, "Gagal menyimpan profil sekolah")
	}

	audit.Log(c, h.db, "update", "school_profiles", 1, old, toResponse(p))
	return httpx.OKWithMessage(c, "Profil sekolah berhasil disimpan", toResponse(p))
}
