package payroll

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"penggajian/api/internal/platform/httpx"
)

var validTemplateEmployeeTypes = map[string]bool{
	"tetap": true, "pppk": true, "honorer": true, "staff_tu": true,
}

func toTemplateResponse(row *templateRow, comps []*TemplateComponentResponse) *TemplateResponse {
	return &TemplateResponse{
		ID: row.ID, Name: row.Name, Description: row.Description,
		EmployeeType: row.EmployeeType, IsDefault: row.IsDefault, IsActive: row.IsActive,
		ComponentCount: row.ComponentCount, Components: comps,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func toTemplateComponentResponse(row *templateComponentRow) *TemplateComponentResponse {
	return &TemplateComponentResponse{
		ID: row.ID, ComponentID: row.ComponentID,
		ComponentCode: row.ComponentCode, ComponentName: row.ComponentName,
		ComponentType: row.ComponentType,
		CalcMode: row.CalcMode, Amount: row.Amount, Note: row.Note,
	}
}

func (s *Service) ListTemplates(ctx context.Context, page, per int64, q string) (*TemplateListResult, error) {
	rows, total, err := s.repo.ListTemplates(ctx, q, page, per)
	if err != nil {
		return nil, err
	}
	resp := make([]*TemplateResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toTemplateResponse(row, nil))
	}
	return &TemplateListResult{Templates: resp, Meta: httpx.NewMeta(page, per, total)}, nil
}

func (s *Service) GetTemplate(ctx context.Context, id int64) (*TemplateResponse, error) {
	row, err := s.repo.FindTemplate(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	compRows, err := s.repo.ListTemplateComponents(ctx, id)
	if err != nil {
		return nil, err
	}
	comps := make([]*TemplateComponentResponse, 0, len(compRows))
	for _, c := range compRows {
		comps = append(comps, toTemplateComponentResponse(c))
	}
	return toTemplateResponse(row, comps), nil
}

// validateTemplate memvalidasi request dan mengembalikan entity + daftar komponen siap simpan.
func (s *Service) validateTemplate(ctx context.Context, req SaveTemplateRequest) (*PayrollTemplate, []*PayrollTemplateComponent, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, nil, errors.New("nama template wajib diisi")
	}
	var employeeType *string
	if req.EmployeeType != nil && strings.TrimSpace(*req.EmployeeType) != "" {
		t := strings.TrimSpace(*req.EmployeeType)
		if !validTemplateEmployeeTypes[t] {
			return nil, nil, errors.New("tipe pegawai tidak valid")
		}
		employeeType = &t
	}

	comps := make([]*PayrollTemplateComponent, 0, len(req.Components))
	seen := map[int64]bool{}
	for i, in := range req.Components {
		if in.ComponentID <= 0 {
			return nil, nil, fmt.Errorf("komponen baris %d wajib dipilih", i+1)
		}
		if seen[in.ComponentID] {
			return nil, nil, fmt.Errorf("komponen baris %d duplikat", i+1)
		}
		seen[in.ComponentID] = true
		if in.Amount < 0 {
			return nil, nil, errors.New("nominal tidak boleh negatif")
		}
		calcMode := in.CalcMode
		if calcMode == "" {
			calcMode = CalcFixed
		}
		if !IsValidCalcMode(calcMode) {
			return nil, nil, fmt.Errorf("mode hitung baris %d tidak valid", i+1)
		}
		if _, err := s.repo.FindComponent(ctx, in.ComponentID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("komponen baris %d tidak ditemukan", i+1)
			}
			return nil, nil, err
		}
		comps = append(comps, &PayrollTemplateComponent{
			ComponentID: in.ComponentID, CalcMode: calcMode, Amount: in.Amount, Note: cleanNote(in.Note),
		})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	t := &PayrollTemplate{
		Name: name, Description: cleanNote(req.Description),
		EmployeeType: employeeType, IsDefault: req.IsDefault, IsActive: isActive,
	}
	return t, comps, nil
}

func (s *Service) CreateTemplate(ctx context.Context, req SaveTemplateRequest) (*TemplateResponse, error) {
	t, comps, err := s.validateTemplate(ctx, req)
	if err != nil {
		return nil, err
	}
	if t.IsDefault {
		if err := s.repo.ClearTemplateDefault(ctx, t.EmployeeType, 0); err != nil {
			return nil, err
		}
	}
	if err := s.repo.CreateTemplateWithComponents(ctx, t, comps); err != nil {
		return nil, err
	}
	return s.GetTemplate(ctx, t.ID)
}

func (s *Service) UpdateTemplate(ctx context.Context, id int64, req SaveTemplateRequest) (*TemplateResponse, error) {
	if _, err := s.repo.FindTemplate(ctx, id); err != nil {
		return nil, mapRepoError(err)
	}
	t, comps, err := s.validateTemplate(ctx, req)
	if err != nil {
		return nil, err
	}
	t.ID = id
	if t.IsDefault {
		if err := s.repo.ClearTemplateDefault(ctx, t.EmployeeType, id); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateTemplateWithComponents(ctx, t, comps); err != nil {
		return nil, err
	}
	return s.GetTemplate(ctx, id)
}

func (s *Service) DeleteTemplate(ctx context.Context, id int64) error {
	if _, err := s.repo.FindTemplate(ctx, id); err != nil {
		return mapRepoError(err)
	}
	return s.repo.SoftDeleteTemplate(ctx, id)
}

// ApplyTemplate menyalin komponen template ke pegawai terpilih (ids eksplisit,
// atau seluruh pegawai aktif bila employee_type diisi; kosong = semua pegawai aktif).
func (s *Service) ApplyTemplate(ctx context.Context, templateID int64, req ApplyTemplateRequest) (*ApplyTemplateResult, error) {
	row, err := s.repo.FindTemplate(ctx, templateID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	if !row.IsActive {
		return nil, errors.New("template tidak aktif")
	}

	ids := req.EmployeeIDs
	if len(ids) == 0 {
		ids, err = s.repo.ListActiveEmployeeIDs(ctx, req.EmployeeType)
		if err != nil {
			return nil, err
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("tidak ada pegawai yang memenuhi kriteria")
	}

	result := &ApplyTemplateResult{TemplateID: templateID, TemplateName: row.Name}
	for _, id := range ids {
		if req.Replace {
			if _, err := s.repo.DeleteEmployeeComponents(ctx, id); err != nil {
				return nil, err
			}
		}
		applied, err := s.repo.ApplyTemplateToEmployee(ctx, templateID, id)
		if err != nil {
			return nil, err
		}
		if applied == 0 && !req.Replace {
			result.Skipped = append(result.Skipped, SkippedEmployee{
				EmployeeID: id, Reason: "semua komponen template sudah dimiliki pegawai",
			})
		}
		result.ComponentsApplied += applied
		result.EmployeesAffected++
	}
	result.Message = fmt.Sprintf("Template %q diterapkan ke %d pegawai (%d komponen ditambahkan)",
		row.Name, result.EmployeesAffected, result.ComponentsApplied)
	return result, nil
}

// ApplyDefaultTemplate dipanggil otomatis saat pegawai baru dibuat.
// Best-effort: kegagalan tidak menggagalkan pembuatan pegawai.
func (s *Service) ApplyDefaultTemplate(ctx context.Context, employeeID int64, employeeType string) error {
	row, err := s.repo.FindDefaultTemplate(ctx, employeeType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // belum ada template default — bukan error
		}
		return err
	}
	_, err = s.repo.ApplyTemplateToEmployee(ctx, row.ID, employeeID)
	return err
}
