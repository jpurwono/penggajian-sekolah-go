package spp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/middleware"
)

type FeeTypeService struct {
	db   *gorm.DB
	repo *FeeTypeRepository
	sppRepo *Repository
}

func NewFeeTypeService(db *gorm.DB, repo *FeeTypeRepository, sppRepo *Repository) *FeeTypeService {
	return &FeeTypeService{db: db, repo: repo, sppRepo: sppRepo}
}

func (s *FeeTypeService) List(ctx context.Context, q string, active *bool) ([]*FeeTypeResponse, error) {
	return s.repo.List(ctx, q, active)
}

func (s *FeeTypeService) validate(ctx context.Context, req SaveFeeTypeRequest, excludeID int64) (*FeeType, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return nil, errors.New("kode & nama jenis biaya wajib diisi")
	}
	if !IsValidFrequency(req.Frequency) {
		return nil, errors.New("frekuensi harus monthly atau annual")
	}
	if exists, err := s.repo.CodeExists(ctx, code, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("kode sudah digunakan")
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &FeeType{
		Code: code, Name: name, Frequency: req.Frequency,
		DefaultAmount: req.DefaultAmount, Description: cleanStr(req.Description),
		IsActive: isActive,
	}, nil
}

func (s *FeeTypeService) Create(ctx context.Context, c echo.Context, req SaveFeeTypeRequest) (*FeeTypeResponse, error) {
	ft, err := s.validate(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, ft); err != nil {
		return nil, err
	}
	audit.Log(c, s.db, "create", "spp_fee_types", ft.ID, nil, ft)
	return s.toResponse(ctx, ft.ID)
}

func (s *FeeTypeService) Update(ctx context.Context, c echo.Context, id int64, req SaveFeeTypeRequest) (*FeeTypeResponse, error) {
	if _, err := s.repo.Find(ctx, id); err != nil {
		return nil, err
	}
	ft, err := s.validate(ctx, req, id)
	if err != nil {
		return nil, err
	}
	ft.ID = id
	if err := s.repo.Update(ctx, ft); err != nil {
		return nil, err
	}
	audit.Log(c, s.db, "update", "spp_fee_types", id, nil, ft)
	return s.toResponse(ctx, id)
}

func (s *FeeTypeService) Delete(ctx context.Context, c echo.Context, id int64) error {
	if _, err := s.repo.Find(ctx, id); err != nil {
		return err
	}
	if count, err := s.repo.CountBillsByType(ctx, id); err != nil {
		return err
	} else if count > 0 {
		return fmt.Errorf("jenis biaya masih dipakai %d tagihan", count)
	}
	audit.Log(c, s.db, "delete", "spp_fee_types", id, nil, nil)
	return s.repo.SoftDelete(ctx, id)
}

func (s *FeeTypeService) toResponse(ctx context.Context, id int64) (*FeeTypeResponse, error) {
	ft, err := s.repo.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	label := "Tahunan"
	if ft.Frequency == FreqMonthly {
		label = "Bulanan"
	}
	return &FeeTypeResponse{
		ID: ft.ID, Code: ft.Code, Name: ft.Name,
		Frequency: ft.Frequency, FrequencyLabel: label,
		DefaultAmount: ft.DefaultAmount, Description: ft.Description,
		IsActive: ft.IsActive, CreatedAt: ft.CreatedAt,
	}, nil
}

// ── Generate Tahunan Multi-Komponen ──

func (s *FeeTypeService) GenerateYearly(ctx context.Context, c echo.Context, req GenerateYearlyRequest) (*GenerateYearlyResult, error) {
	if req.Year < 2000 || req.Year > 2100 {
		return nil, errors.New("tahun tidak valid")
	}
	if len(req.FeeTypeIDs) == 0 {
		return nil, errors.New("pilih minimal satu jenis biaya")
	}

	className := ""
	if req.ClassName != nil {
		className = strings.TrimSpace(*req.ClassName)
	}
	students, err := s.sppRepo.ListActiveStudents(ctx, className)
	if err != nil {
		return nil, err
	}
	if len(students) == 0 {
		return nil, errors.New("belum ada siswa aktif")
	}

	actorID := middleware.CurrentUserID(c)
	var totalCreated, totalSkipped int
	details := []GenerateDetail{}

	for _, ftID := range req.FeeTypeIDs {
		ft, err := s.repo.Find(ctx, ftID)
		if err != nil {
			return nil, fmt.Errorf("jenis biaya id %d tidak ditemukan", ftID)
		}

		amount := ft.DefaultAmount
		if custom, ok := req.CustomAmounts[ftID]; ok && custom > 0 {
			amount = custom
		}
		if amount <= 0 {
			return nil, fmt.Errorf("nominal untuk %s belum diisi (set default atau custom)", ft.Name)
		}

		created, skipped := 0, 0

		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			months := []int{}
			if ft.Frequency == FreqMonthly {
				for m := 1; m <= 12; m++ {
					months = append(months, m)
				}
			}

			processMonth := func(month *int) error {
				year := req.Year
				for _, student := range students {
					exists, err := s.repo.BillExists(ctx, student.ID, ft.ID, month, year)
					if err != nil {
						return err
					}
					if exists {
						skipped++
						continue
					}

					name := ft.Name
					billNo, err := s.sppRepo.NextBillNo(ctx, tx, time.Now())
					if err != nil {
						return err
					}

					bill := &Bill{
						BillNo: billNo, StudentID: student.ID,
						FeeTypeID: &ft.ID, Name: name,
						PeriodMonth: month, PeriodYear: &year,
						Amount: amount, Status: BillUnpaid,
						DueDate: parseDatePtr(req.DueDate),
						CreatedBy: &actorID,
					}
					if err := tx.Create(bill).Error; err != nil {
						return err
					}
					created++
				}
				return nil
			}

			if len(months) > 0 {
				for _, m := range months {
					mm := m
					if err := processMonth(&mm); err != nil {
						return err
					}
				}
			} else {
				if err := processMonth(nil); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("generate %s: %w", ft.Name, err)
		}

		totalCreated += created
		totalSkipped += skipped
		details = append(details, GenerateDetail{
			FeeTypeCode: ft.Code, FeeTypeName: ft.Name,
			Created: created, Skipped: skipped, Amount: amount,
		})
	}

	audit.Log(c, s.db, "generate", "spp_bills", 0, nil, map[string]any{
		"year": req.Year, "total_created": totalCreated, "details": len(details),
	})

	freqLabel := map[string]string{"monthly": "12 bulan", "annual": "1x"}
	_ = freqLabel

	return &GenerateYearlyResult{
		TotalCreated: totalCreated, TotalSkipped: totalSkipped,
		Details: details,
		Message: fmt.Sprintf("%d tagihan dibuat, %d dilewati", totalCreated, totalSkipped),
	}, nil
}

func parseDatePtr(s *string) *time.Time {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(*s))
	if err != nil {
		return nil
	}
	return &t
}
