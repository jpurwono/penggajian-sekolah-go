package payroll

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
)

var ErrNotFound = errors.New("data tidak ditemukan")

var codePattern = regexp.MustCompile(`^[A-Z0-9_]{2,20}$`)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func mapRepoError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func toComponentResponse(c *PayComponent) *ComponentResponse {
	return &ComponentResponse{
		ID: c.ID, Code: c.Code, Name: c.Name, Type: c.Type,
		Description: c.Description, IsActive: c.IsActive, CreatedAt: c.CreatedAt,
	}
}

func (s *Service) ListComponents(ctx context.Context, f ComponentFilter) (*ComponentListResult, error) {
	list, total, err := s.repo.ListComponents(ctx, f)
	if err != nil {
		return nil, err
	}
	resp := make([]*ComponentResponse, 0, len(list))
	for _, c := range list {
		resp = append(resp, toComponentResponse(c))
	}
	return &ComponentListResult{Components: resp, Meta: httpx.NewMeta(f.Page, f.PerPage, total)}, nil
}

func (s *Service) GetComponent(ctx context.Context, id int64) (*ComponentResponse, error) {
	c, err := s.repo.FindComponent(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	return toComponentResponse(c), nil
}

func (s *Service) validateComponent(ctx context.Context, req SaveComponentRequest, excludeID int64) (*PayComponent, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if !codePattern.MatchString(code) {
		return nil, errors.New("kode komponen 2-20 karakter (huruf besar, angka, underscore)")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("nama komponen wajib diisi")
	}
	if !IsValidComponentType(req.Type) {
		return nil, errors.New("tipe komponen harus earning atau deduction")
	}
	if exists, err := s.repo.ComponentCodeExists(ctx, code, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("kode komponen sudah digunakan")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var description *string
	if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
		description = req.Description
	}

	return &PayComponent{
		Code: code, Name: strings.TrimSpace(req.Name), Type: req.Type,
		Description: description, IsActive: isActive,
	}, nil
}

func (s *Service) CreateComponent(ctx context.Context, req SaveComponentRequest) (*ComponentResponse, error) {
	comp, err := s.validateComponent(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateComponent(ctx, comp); err != nil {
		return nil, err
	}
	return s.GetComponent(ctx, comp.ID)
}

func (s *Service) UpdateComponent(ctx context.Context, id int64, req SaveComponentRequest) (*ComponentResponse, error) {
	if _, err := s.repo.FindComponent(ctx, id); err != nil {
		return nil, mapRepoError(err)
	}
	comp, err := s.validateComponent(ctx, req, id)
	if err != nil {
		return nil, err
	}
	comp.ID = id
	if err := s.repo.UpdateComponent(ctx, comp); err != nil {
		return nil, err
	}
	return s.GetComponent(ctx, id)
}

func (s *Service) DeleteComponent(ctx context.Context, id int64) error {
	if _, err := s.repo.FindComponent(ctx, id); err != nil {
		return mapRepoError(err)
	}
	count, err := s.repo.CountComponentAssignments(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("komponen masih dipakai oleh %d pegawai, hapus dulu komponen pegawainya", count)
	}
	return s.repo.SoftDeleteComponent(ctx, id)
}

func toEmployeeComponentResponse(row *EmployeeComponentRow) *EmployeeComponentResponse {
	return &EmployeeComponentResponse{
		ID: row.ID, EmployeeID: row.EmployeeID, EmployeeName: row.EmployeeName,
		ComponentID: row.ComponentID, ComponentCode: row.ComponentCode,
		ComponentName: row.ComponentName, ComponentType: row.ComponentType,
		CalcMode: row.CalcMode, Amount: row.Amount, Note: row.Note, CreatedAt: row.CreatedAt,
	}
}

func (s *Service) ListEmployeeComponents(ctx context.Context, employeeID, page, per int64) (*EmployeeComponentListResult, error) {
	rows, total, err := s.repo.ListEmployeeComponents(ctx, employeeID, page, per)
	if err != nil {
		return nil, err
	}
	resp := make([]*EmployeeComponentResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toEmployeeComponentResponse(row))
	}
	return &EmployeeComponentListResult{EmployeeComponents: resp, Meta: httpx.NewMeta(page, per, total)}, nil
}

func (s *Service) AddEmployeeComponent(ctx context.Context, req SaveEmployeeComponentRequest) (*EmployeeComponentResponse, error) {
	calcMode, err := s.validateEmployeeComponent(ctx, req)
	if err != nil {
		return nil, err
	}

	ec := &EmployeeComponent{
		EmployeeID: req.EmployeeID, ComponentID: req.ComponentID,
		CalcMode: calcMode, Amount: req.Amount, Note: cleanNote(req.Note),
	}
	if err := s.repo.CreateEmployeeComponent(ctx, ec); err != nil {
		return nil, err
	}
	row, err := s.repo.FindEmployeeComponent(ctx, ec.ID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	return toEmployeeComponentResponse(row), nil
}

func (s *Service) validateEmployeeComponent(ctx context.Context, req SaveEmployeeComponentRequest) (string, error) {
	if req.EmployeeID <= 0 {
		return "", errors.New("pegawai wajib dipilih")
	}
	if req.ComponentID <= 0 {
		return "", errors.New("komponen wajib dipilih")
	}
	if req.Amount < 0 {
		return "", errors.New("nominal tidak boleh negatif")
	}
	calcMode := req.CalcMode
	if calcMode == "" {
		calcMode = CalcFixed
	}
	if !IsValidCalcMode(calcMode) {
		return "", errors.New("mode hitung tidak valid")
	}
	if _, err := s.repo.FindComponent(ctx, req.ComponentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("komponen tidak ditemukan")
		}
		return "", err
	}
	if exists, err := s.repo.EmployeeComponentExists(ctx, req.EmployeeID, req.ComponentID, 0); err != nil {
		return "", err
	} else if exists {
		return "", errors.New("komponen ini sudah ada pada pegawai tersebut")
	}
	return calcMode, nil
}

func cleanNote(note *string) *string {
	if note == nil || strings.TrimSpace(*note) == "" {
		return nil
	}
	return note
}

func (s *Service) UpdateEmployeeComponent(ctx context.Context, id int64, req SaveEmployeeComponentRequest) (*EmployeeComponentResponse, error) {
	if _, err := s.repo.FindEmployeeComponent(ctx, id); err != nil {
		return nil, mapRepoError(err)
	}
	if req.Amount < 0 {
		return nil, errors.New("nominal tidak boleh negatif")
	}
	calcMode := req.CalcMode
	if calcMode == "" {
		calcMode = CalcFixed
	}
	if !IsValidCalcMode(calcMode) {
		return nil, errors.New("mode hitung tidak valid")
	}
	ec := &EmployeeComponent{ID: id, CalcMode: calcMode, Amount: req.Amount, Note: cleanNote(req.Note)}
	if err := s.repo.UpdateEmployeeComponent(ctx, ec); err != nil {
		return nil, err
	}
	row, err := s.repo.FindEmployeeComponent(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	return toEmployeeComponentResponse(row), nil
}

func (s *Service) DeleteEmployeeComponent(ctx context.Context, id int64) error {
	if _, err := s.repo.FindEmployeeComponent(ctx, id); err != nil {
		return mapRepoError(err)
	}
	return s.repo.DeleteEmployeeComponent(ctx, id)
}

func toPeriodResponse(row *periodRow) *PeriodResponse {
	return &PeriodResponse{
		ID: row.ID, Name: row.Name, PeriodMonth: row.PeriodMonth, PeriodYear: row.PeriodYear,
		Status: row.Status, PaidAt: row.PaidAt, Note: row.Note,
		EffectiveWeeks: row.EffectiveWeeks,
		SlipCount: row.SlipCount, TotalEarning: row.TotalEarning,
		TotalDeduction: row.TotalDeduction, TotalTakeHome: row.TotalTakeHome,
		CreatedAt: row.CreatedAt,
	}
}

func (s *Service) ListPeriods(ctx context.Context, f PeriodFilter) (*PeriodListResult, error) {
	rows, total, err := s.repo.ListPeriods(ctx, f)
	if err != nil {
		return nil, err
	}
	resp := make([]*PeriodResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toPeriodResponse(row))
	}
	return &PeriodListResult{Periods: resp, Meta: httpx.NewMeta(f.Page, f.PerPage, total)}, nil
}

func (s *Service) GetPeriod(ctx context.Context, id int64) (*PeriodResponse, error) {
	row, err := s.repo.FindPeriodRow(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	return toPeriodResponse(row), nil
}

func (s *Service) validatePeriod(ctx context.Context, req SavePeriodRequest, excludeID int64) (*PayrollPeriod, error) {
	if req.PeriodMonth < 1 || req.PeriodMonth > 12 {
		return nil, errors.New("bulan harus 1-12")
	}
	if req.PeriodYear < 1900 || req.PeriodYear > 2100 {
		return nil, errors.New("tahun tidak valid")
	}
	if exists, err := s.repo.PeriodYearMonthExists(ctx, req.PeriodYear, req.PeriodMonth, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("periode untuk bulan & tahun tersebut sudah ada")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("Gaji %s %d", MonthNames[req.PeriodMonth], req.PeriodYear)
	}
	return &PayrollPeriod{
		Name: name, PeriodMonth: req.PeriodMonth, PeriodYear: req.PeriodYear,
		Note: cleanNote(req.Note),
	}, nil
}

func (s *Service) CreatePeriod(ctx context.Context, req SavePeriodRequest, actorID int64) (*PeriodResponse, error) {
	period, err := s.validatePeriod(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	workDays, err := s.repo.SchoolWorkingDays(ctx)
	if err != nil {
		return nil, err
	}
	eff := EffectiveWeeks(req.PeriodYear, req.PeriodMonth, workDays)
	period.Status = StatusDraft
	period.CreatedBy = &actorID
	period.EffectiveWeeks = &eff
	if err := s.repo.CreatePeriod(ctx, period); err != nil {
		return nil, err
	}
	return s.GetPeriod(ctx, period.ID)
}

func (s *Service) UpdatePeriod(ctx context.Context, id int64, req SavePeriodRequest) (*PeriodResponse, error) {
	period, err := s.repo.FindPeriod(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	if period.Status != StatusDraft {
		return nil, errors.New("hanya periode draft yang dapat diubah")
	}
	updated, err := s.validatePeriod(ctx, req, id)
	if err != nil {
		return nil, err
	}
	workDays, err := s.repo.SchoolWorkingDays(ctx)
	if err != nil {
		return nil, err
	}
	eff := EffectiveWeeks(req.PeriodYear, req.PeriodMonth, workDays)
	period.Name = updated.Name
	period.Note = updated.Note
	period.PeriodMonth = updated.PeriodMonth
	period.PeriodYear = updated.PeriodYear
	period.EffectiveWeeks = &eff
	if err := s.repo.UpdatePeriod(ctx, period); err != nil {
		return nil, err
	}
	return s.GetPeriod(ctx, id)
}

func (s *Service) DeletePeriod(ctx context.Context, id int64) error {
	period, err := s.repo.FindPeriod(ctx, id)
	if err != nil {
		return mapRepoError(err)
	}
	if period.Status != StatusDraft {
		return nil
	}
	return s.repo.DeleteDraftPeriod(ctx, id)
}

func (s *Service) transitionPeriod(ctx context.Context, id int64, from []string, to string, setPaidAt bool) (*PeriodResponse, error) {
	period, err := s.repo.FindPeriod(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	allowed := false
	for _, f := range from {
		if period.Status == f {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("periode berstatus %s tidak dapat diubah menjadi %s", period.Status, to)
	}
	if err := s.repo.UpdatePeriodStatus(ctx, id, to, setPaidAt); err != nil {
		return nil, err
	}
	return s.GetPeriod(ctx, id)
}

func (s *Service) ApprovePeriod(ctx context.Context, id int64) (*PeriodResponse, error) {
	return s.transitionPeriod(ctx, id, []string{StatusDraft}, StatusApproved, false)
}

func (s *Service) PayPeriod(ctx context.Context, id int64) (*PeriodResponse, error) {
	return s.transitionPeriod(ctx, id, []string{StatusApproved}, StatusPaid, true)
}

func (s *Service) CancelPeriod(ctx context.Context, id int64) (*PeriodResponse, error) {
	return s.transitionPeriod(ctx, id, []string{StatusDraft, StatusApproved}, StatusCancelled, false)
}
