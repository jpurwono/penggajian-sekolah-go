package attendance

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/middleware"
)

var ErrNotFound = errors.New("data absen tidak ditemukan")

type Service struct {
	db   *gorm.DB
	repo *Repository
}

func NewService(db *gorm.DB, repo *Repository) *Service {
	return &Service{db: db, repo: repo}
}

func (s *Service) CheckIn(ctx context.Context, c echo.Context, req CheckInRequest) (*CheckInResult, error) {
	user := middleware.CurrentUser(c)
	if user == nil || user.EmployeeID == nil || *user.EmployeeID <= 0 {
		return nil, errors.New("akun belum ditautkan ke data pegawai — hubungi admin")
	}
	employeeID := *user.EmployeeID

	if !IsValidJenis(req.Jenis) {
		return nil, errors.New("jenis absen harus 'hadir' atau 'kelas'")
	}
	if req.Latitude == 0 && req.Longitude == 0 {
		return nil, errors.New("koordinat GPS tidak valid")
	}

	recordedAt := time.Now()
	if strings.TrimSpace(req.RecordedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.RecordedAt))
		if err != nil {
			return nil, errors.New("format waktu harus RFC3339")
		}
		recordedAt = parsed
	}

	// Validasi hadir: cegah duplikat hari yang sama
	if req.Jenis == JenisHadir {
		already, err := s.repo.HadirToday(ctx, employeeID, recordedAt)
		if err != nil {
			return nil, err
		}
		if already {
			return nil, errors.New("absen kehadiran hari ini sudah tercatat")
		}
	}

	// Validasi kelas: teaching_id wajib dan harus milik guru ini
	if req.Jenis == JenisKelas {
		if req.TeachingID == nil || *req.TeachingID <= 0 {
			return nil, errors.New("absen kelas wajib memilih jadwal (teaching_id)")
		}
		var owner int64
		err := s.db.WithContext(ctx).Raw(
			`SELECT employee_id FROM teaching_assignments WHERE id = ?`, *req.TeachingID,
		).Scan(&owner).Error
		if err != nil {
			return nil, err
		}
		if owner == 0 || owner != employeeID {
			return nil, errors.New("jadwal mengajar tidak ditemukan / bukan milik Anda")
		}
	}

	// Geofence: hitung jarak ke sekolah
	schLat, schLon, radiusM, err := s.repo.SchoolGeo(ctx)
	if err != nil {
		return nil, err
	}
	var distance float64
	var distancePtr *float64
	if schLat != 0 && schLon != 0 {
		distance = HaversineMeters(schLat, schLon, req.Latitude, req.Longitude)
		distancePtr = &distance
	}

	rec := &AttendanceRecord{
		EmployeeID: employeeID,
		Jenis:      req.Jenis,
		TeachingID: req.TeachingID,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		AccuracyM:  req.AccuracyM,
		IsMock:     req.IsMock,
		DistanceM:  distancePtr,
		DeviceInfo: &req.DeviceInfo,
		RecordedAt: recordedAt,
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, err
	}

	audit.Log(c, s.db, "create", "attendance_records", rec.ID, nil, map[string]any{
		"jenis": req.Jenis, "distance_m": distance, "is_mock": req.IsMock,
	})

	row, err := s.repo.Find(ctx, rec.ID)
	if err != nil {
		return nil, err
	}

	within := schLat == 0 || distance <= float64(radiusM)
	msg := "Absen berhasil"
	if !within {
		msg = "Absen tercatat (DI LUAR area sekolah — jarak " + formatM(distance) + ")"
	}

	return &CheckInResult{
		Record:   rowToResponse(row),
		Distance: distance,
		Within:   within,
		Message:  msg,
	}, nil
}

func formatM(m float64) string {
	return fmt.Sprintf("%.0f m", m)
}

func (s *Service) List(ctx context.Context, f Filter, page, per int64) ([]*AttendanceResponse, *httpx.Meta, error) {
	rows, total, err := s.repo.List(ctx, f, per, (page-1)*per)
	if err != nil {
		return nil, nil, err
	}
	return rows, httpx.NewMeta(page, per, total), nil
}

func (s *Service) Get(ctx context.Context, id int64) (*AttendanceResponse, error) {
	row, err := s.repo.Find(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return rowToResponse(row), nil
}

func (s *Service) DailySummary(ctx context.Context, date string) (*DailySummary, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	rows, err := s.repo.DailySummary(ctx, date)
	if err != nil {
		return nil, err
	}
	return &DailySummary{Date: date, Rows: rows}, nil
}
