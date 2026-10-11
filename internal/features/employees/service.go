package employees

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
)

var ErrNotFound = errors.New("pegawai tidak ditemukan")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func parseDate(value *string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, strings.TrimSpace(*value))
	if err != nil {
		return nil, errors.New("format tanggal harus YYYY-MM-DD")
	}
	return &t, nil
}

func strPtr(v string) *string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *Service) validate(ctx context.Context, req SaveEmployeeRequest, excludeID int64) (*Employee, error) {
	employeeNo := strings.TrimSpace(req.EmployeeNo)
	if employeeNo == "" {
		return nil, errors.New("nomor pegawai wajib diisi")
	}
	if len(employeeNo) > 30 {
		return nil, errors.New("nomor pegawai maksimal 30 karakter")
	}
	if strings.TrimSpace(req.FullName) == "" {
		return nil, errors.New("nama lengkap wajib diisi")
	}
	if !IsValidEmployeeType(req.EmployeeType) {
		return nil, errors.New("jenis pegawai tidak valid (tetap/pppk/honorer/staff_tu)")
	}
	if req.BaseSalary < 0 {
		return nil, errors.New("gaji pokok tidak boleh negatif")
	}
	if req.Gender != nil && *req.Gender != "" && *req.Gender != "L" && *req.Gender != "P" {
		return nil, errors.New("jenis kelamin harus L atau P")
	}

	joinDate, err := parseDate(&req.JoinDate)
	if err != nil || joinDate == nil {
		return nil, errors.New("tanggal masuk wajib diisi dengan format YYYY-MM-DD")
	}
	birthDate, err := parseDate(req.BirthDate)
	if err != nil {
		return nil, err
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return nil, err
	}
	if endDate != nil && endDate.Before(*joinDate) {
		return nil, errors.New("tanggal keluar tidak boleh sebelum tanggal masuk")
	}

	if exists, err := s.repo.EmployeeNoExists(ctx, employeeNo, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("nomor pegawai sudah digunakan")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return &Employee{
		EmployeeNo:   employeeNo,
		FullName:     strings.TrimSpace(req.FullName),
		Gender:       normalizeGender(req.Gender),
		BirthDate:    birthDate,
		Address:      strPtr(deref(req.Address)),
		Phone:        strPtr(deref(req.Phone)),
		Email:        strPtr(deref(req.Email)),
		EmployeeType: req.EmployeeType,
		JobTitle:     strPtr(deref(req.JobTitle)),
		JoinDate:     *joinDate,
		EndDate:      endDate,
		BaseSalary:   req.BaseSalary,
		BankName:     strPtr(deref(req.BankName)),
		BankAccount:  strPtr(deref(req.BankAccount)),
		PhotoURL:     strPtr(deref(req.PhotoURL)),
		IsActive:     isActive,
	}, nil
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func normalizeGender(p *string) *string {
	if p == nil || strings.TrimSpace(*p) == "" {
		return nil
	}
	v := strings.ToUpper(strings.TrimSpace(*p))
	return &v
}

func (s *Service) List(ctx context.Context, f ListFilter) (*EmployeeListResult, error) {
	list, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	resp := make([]*EmployeeResponse, 0, len(list))
	for _, e := range list {
		resp = append(resp, toResponse(e))
	}
	return &EmployeeListResult{Employees: resp, Meta: httpx.NewMeta(f.Page, f.PerPage, total)}, nil
}

func (s *Service) Get(ctx context.Context, id int64) (*EmployeeResponse, error) {
	e, err := s.repo.Find(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toResponse(e), nil
}

func (s *Service) Create(ctx context.Context, req SaveEmployeeRequest) (*EmployeeResponse, error) {
	employee, err := s.validate(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, employee); err != nil {
		return nil, err
	}
	return s.Get(ctx, employee.ID)
}

func (s *Service) Update(ctx context.Context, id int64, req SaveEmployeeRequest) (*EmployeeResponse, error) {
	if _, err := s.repo.Find(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	employee, err := s.validate(ctx, req, id)
	if err != nil {
		return nil, err
	}
	employee.ID = id

	if err := s.repo.Update(ctx, employee); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.Find(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) CountActive(ctx context.Context) (int64, error) {
	return s.repo.CountActive(ctx)
}
