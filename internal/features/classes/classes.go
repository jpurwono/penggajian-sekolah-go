package classes

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
	mw "penggajian/server/internal/platform/middleware"
)

type Class struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Level       *string   `gorm:"column:level" json:"level"`
	Description *string   `gorm:"column:description" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Class) TableName() string { return "classes" }

type ClassResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Level       *string   `json:"level"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type SaveClassRequest struct {
	Name        string  `json:"name"`
	Level       *string `json:"level"`
	Description *string `json:"description"`
}

func toResponse(c *Class) *ClassResponse {
	return &ClassResponse{ID: c.ID, Name: c.Name, Level: c.Level, Description: c.Description, CreatedAt: c.CreatedAt}
}

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) validate(ctx context.Context, req SaveClassRequest, excludeID int64) (*Class, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("nama kelas wajib diisi")
	}
	var count int64
	q := s.db.WithContext(ctx).Unscoped().Model(&Class{}).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("nama kelas sudah digunakan")
	}
	clean := func(v *string) *string {
		if v == nil || strings.TrimSpace(*v) == "" {
			return nil
		}
		return v
	}
	return &Class{Name: name, Level: clean(req.Level), Description: clean(req.Description)}, nil
}

func (s *Service) List(ctx context.Context, q string, limit, offset int64) ([]*ClassResponse, int64, error) {
	query := s.db.WithContext(ctx).Model(&Class{})
	if q != "" {
		query = query.Where("name ILIKE ? OR COALESCE(level, '') ILIKE ?", "%"+q+"%", "%"+q+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*Class
	if err := query.Order("name ASC").Offset(int(offset)).Limit(int(limit)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	resp := make([]*ClassResponse, 0, len(list))
	for _, c := range list {
		resp = append(resp, toResponse(c))
	}
	return resp, total, nil
}

func (s *Service) Create(ctx context.Context, req SaveClassRequest) (*ClassResponse, error) {
	c, err := s.validate(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Create(c).Error; err != nil {
		return nil, err
	}
	return toResponse(c), nil
}

func (s *Service) Update(ctx context.Context, id int64, req SaveClassRequest) (*ClassResponse, error) {
	var existing Class
	if err := s.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kelas tidak ditemukan")
		}
		return nil, err
	}
	c, err := s.validate(ctx, req, id)
	if err != nil {
		return nil, err
	}
	c.ID = id
	if err := s.db.WithContext(ctx).Model(c).
		Select("name", "level", "description").Updates(c).Error; err != nil {
		return nil, err
	}
	return toResponse(c), nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	var count int64
	if err := s.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM employee_teachings WHERE class_id = ?`, id,
	).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("kelas masih dipakai jadwal mengajar pegawai")
	}
	res := s.db.WithContext(ctx).Delete(&Class{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("kelas tidak ditemukan")
	}
	return nil
}

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func mapErr(c echo.Context, err error) error {
	return httpx.ErrBadRequest(c, err.Error())
}

func (h *Handler) List(c echo.Context) error {
	page, per, q := httpx.ParseListParams(c)
	data, total, err := h.svc.List(c.Request().Context(), q, per, (page-1)*per)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OKWithMeta(c, data, httpx.NewMeta(page, per, total))
}

func (h *Handler) Create(c echo.Context) error {
	var req SaveClassRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.Create(c.Request().Context(), req)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, resp)
}

func (h *Handler) Update(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	var req SaveClassRequest
	if err := c.Bind(&req); err != nil {
		return httpx.ErrBadRequest(c, "Format permintaan tidak valid")
	}
	resp, err := h.svc.Update(c.Request().Context(), id, req)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, resp)
}

func (h *Handler) Delete(c echo.Context) error {
	id, _ := httpx.ParseID(c, "id")
	if id <= 0 {
		return httpx.ErrBadRequest(c, "ID tidak valid")
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return mapErr(c, err)
	}
	return httpx.OKWithMessage(c, "Kelas berhasil dihapus", nil)
}

// Register mendaftarkan routes fitur classes ke group /api/v1.
func Register(g *echo.Group, db *gorm.DB, authMW echo.MiddlewareFunc) {
	h := NewHandler(NewService(db))

	grp := g.Group("/classes", authMW, mw.RequirePermission("pegawai.read"))
	grp.GET("", h.List)
	grp.POST("", h.Create, mw.RequirePermission("pegawai.create"))
	grp.PUT("/:id", h.Update, mw.RequirePermission("pegawai.update"))
	grp.DELETE("/:id", h.Delete, mw.RequirePermission("pegawai.delete"))
}
