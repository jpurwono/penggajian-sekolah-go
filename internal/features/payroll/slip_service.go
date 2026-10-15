package payroll

import (
	"context"
	"errors"
	"fmt"
	"math"

	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
)

type SlipService struct {
	db   *gorm.DB
	repo *Repository
	slip *SlipRepository
}

func NewSlipService(db *gorm.DB, repo *Repository, slipRepo *SlipRepository) *SlipService {
	return &SlipService{db: db, repo: repo, slip: slipRepo}
}

// qtyFor menghitung kuantitas sesuai mode; nil bila data jam belum tersedia.
// Untuk jam hadir: Jml Hadir = round(Jml Jam × mingguEfektif) − Jam Tidak Hadir.
// Untuk jam minggu ke-1: Jml Jam × fraksi hari kerja pekan pertama (kalender).
func qtyFor(mode string, jamMengajar, jamTidakHadir, hariHadir *int, cal periodCalendar) *int {
	switch mode {
	case CalcPerJamMengajar:
		return jamMengajar
	case CalcPerHariHadir:
		return hariHadir
	case CalcPerJamHadir:
		if jamMengajar == nil {
			return nil
		}
		tdk := 0
		if jamTidakHadir != nil {
			tdk = *jamTidakHadir
		}
		q := int(math.Round(float64(*jamMengajar)*cal.EffectiveWeeks)) - tdk
		return &q
	case CalcPerJamMinggu1:
		if jamMengajar == nil || cal.WorkDays == 0 {
			return nil
		}
		q := int(math.Round(float64(*jamMengajar) * float64(cal.FirstWeekDays) / float64(cal.WorkDays)))
		return &q
	}
	return nil
}

func autoNote(qty int, tarif int64) *string {
	s := fmt.Sprintf("%d × %d", qty, tarif)
	return &s
}

// Generate membuat ulang seluruh slip untuk periode (hanya draft):
// gaji pokok (GP) + seluruh komponen recurring pegawai, disalin sebagai snapshot.
// Komponen berbasis tarif dihitung dari data jam (dipertahankan saat generate ulang).
func (s *SlipService) Generate(ctx context.Context, periodID int64) (*GenerateResult, error) {
	period, err := s.repo.FindPeriod(ctx, periodID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if period.Status != StatusDraft {
		return nil, errors.New("slip hanya bisa digenerate untuk periode berstatus draft")
	}

	gp, gpErr := s.repo.FindComponentByCode(ctx, ComponentCodeGajiPokok)
	if gpErr != nil && !errors.Is(gpErr, gorm.ErrRecordNotFound) {
		return nil, gpErr
	}

	employees, err := s.slip.ListActiveEmployees(ctx)
	if err != nil {
		return nil, err
	}
	if len(employees) == 0 {
		return nil, errors.New("belum ada pegawai aktif untuk digenerate")
	}
	for _, emp := range employees {
		if emp.BaseSalary > 0 && gpErr != nil {
			return nil, errors.New("komponen GP (Gaji Pokok) belum ada — buat dulu di master komponen")
		}
	}

	assignments, err := s.slip.ListAssignments(ctx)
	if err != nil {
		return nil, err
	}
	assignmentsByEmployee := map[int64][]*assignmentRow{}
	for _, a := range assignments {
		assignmentsByEmployee[a.EmployeeID] = append(assignmentsByEmployee[a.EmployeeID], a)
	}

	workDays, wdErr := s.repo.SchoolWorkingDays(ctx)
	if wdErr != nil {
		return nil, wdErr
	}
	cal := calendarFor(period, workDays)

	var count int
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		oldJam, err := s.slip.ListPeriodJam(ctx, tx, periodID)
		if err != nil {
			return err
		}
		attStats, err := s.slip.AttendanceStats(ctx, tx, period)
		if err != nil {
			return err
		}
		if err := s.slip.DeletePeriodSlips(ctx, tx, periodID); err != nil {
			return err
		}

		seq := 0
		for _, emp := range employees {
			seq++
			jam := mergeJam(oldJam[emp.ID], attStats[emp.ID])

			slip := &Payslip{
				PeriodID:       periodID,
				EmployeeID:     emp.ID,
				SlipNo:         s.slip.NextSlipNo(period, seq),
				JamMengajar:    jamJam(jam, "mengajar"),
				JamTidakHadir:  jamJam(jam, "tidak"),
				HariHadir:      jamJam(jam, "hari"),
			}

			items := []*PayslipItem{}
			var totalEarning, totalDeduction int64
			if emp.BaseSalary > 0 {
				items = append(items, &PayslipItem{
					ComponentID:  &gp.ID,
					CodeSnapshot: gp.Code,
					NameSnapshot: gp.Name,
					Type:         TypeEarning,
					Amount:       emp.BaseSalary,
					IsAuto:       true,
				})
				totalEarning += emp.BaseSalary
			}

			for _, a := range assignmentsByEmployee[emp.ID] {
				var jamMengajar, jamTidakHadir, hariHadir *int
				if jam != nil {
					jamMengajar, jamTidakHadir, hariHadir = jam.JamMengajar, jam.JamTidakHadir, jam.HariHadir
				}
				amount, note, _ := resolveAssignmentAmount(a, jamMengajar, jamTidakHadir, hariHadir, cal)

				componentID := a.ComponentID
				items = append(items, &PayslipItem{
					ComponentID:  &componentID,
					CodeSnapshot: a.Code,
					NameSnapshot: a.Name,
					Type:         a.Type,
					Amount:       amount,
					Note:         note,
					IsAuto:       true,
				})
				if a.Type == TypeEarning {
					totalEarning += amount
				} else {
					totalDeduction += amount
				}
			}

			slip.TotalEarning = totalEarning
			slip.TotalDeduction = totalDeduction
			slip.TakeHomePay = totalEarning - totalDeduction

			if err := s.slip.CreateSlipWithItems(tx, slip, items); err != nil {
				return err
			}
			count++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &GenerateResult{
		PeriodID:  periodID,
		SlipCount: count,
		Message:   fmt.Sprintf("%d slip berhasil digenerate", count),
	}, nil
}

// mergeJam mengisi jam yang masih kosong (manual belum diisi) dari data absensi otomatis.
func mergeJam(manual *slipJamRow, auto *attendanceStatRow) *slipJamRow {
	if auto == nil {
		return manual
	}
	res := &slipJamRow{}
	if manual != nil {
		res = &slipJamRow{JamMengajar: manual.JamMengajar, JamTidakHadir: manual.JamTidakHadir, HariHadir: manual.HariHadir}
	}
	if res.JamMengajar == nil && auto.JamMengajar > 0 {
		v := auto.JamMengajar
		res.JamMengajar = &v
	}
	if res.HariHadir == nil && auto.HariHadir > 0 {
		v := auto.HariHadir
		res.HariHadir = &v
	}
	return res
}

// ImportAttendance menarik jam mengajar & hari hadir dari attendance_records
// (aplikasi Android) ke seluruh slip periode, lalu menghitung ulang slip.
// Jam tidak hadir tidak ditimpa. Periode harus draft.
func (s *SlipService) ImportAttendance(ctx context.Context, periodID int64) (*GenerateResult, error) {
	period, err := s.repo.FindPeriod(ctx, periodID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if period.Status != StatusDraft {
		return nil, errors.New("hanya periode draft yang bisa menarik data absensi")
	}

	oldJam, err := s.slip.ListPeriodJam(ctx, s.db, periodID)
	if err != nil {
		return nil, err
	}
	if len(oldJam) == 0 {
		return nil, errors.New("belum ada slip — generate slip terlebih dahulu")
	}

	stats, err := s.slip.AttendanceStats(ctx, s.db, period)
	if err != nil {
		return nil, err
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for empID, st := range stats {
			if err := s.slip.SetJamFromImport(ctx, tx, periodID, empID, st.JamMengajar, st.HariHadir); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	res, err := s.Generate(ctx, periodID)
	if err != nil {
		return nil, err
	}
	res.Message = fmt.Sprintf("Absensi %d pegawai ditarik; %d slip dihitung ulang", len(stats), res.SlipCount)
	return res, nil
}

func jamJam(jam *slipJamRow, field string) *int {
	if jam == nil {
		return nil
	}
	switch field {
	case "mengajar":
		return jam.JamMengajar
	case "tidak":
		return jam.JamTidakHadir
	case "hari":
		return jam.HariHadir
	}
	return nil
}

func resolveAssignmentAmount(a *assignmentRow, jamMengajar, jamTidakHadir, hariHadir *int, cal periodCalendar) (int64, *string, bool) {
	if a.CalcMode == "" || a.CalcMode == CalcFixed {
		return a.Amount, a.Note, false
	}
	qty := qtyFor(a.CalcMode, jamMengajar, jamTidakHadir, hariHadir, cal)
	if qty == nil {
		note := autoNote(0, a.Amount)
		return 0, note, true
	}
	return a.Amount * int64(*qty), autoNote(*qty, a.Amount), true
}

func toPayslipResponse(row *payslipRow) *PayslipResponse {
	resp := &PayslipResponse{
		ID: row.ID, PeriodID: row.PeriodID, SlipNo: row.SlipNo,
		EmployeeID: row.EmployeeID, EmployeeNo: row.EmployeeNo,
		EmployeeName: row.EmployeeName, JobTitle: derefString(row.JobTitle),
		JamMengajar: row.JamMengajar, JamTidakHadir: row.JamTidakHadir, HariHadir: row.HariHadir,
		TotalEarning: row.TotalEarning, TotalDeduction: row.TotalDeduction,
		TakeHomePay: row.TakeHomePay, Note: row.Note,
	}
	if row.JamMengajar != nil {
		tdk := 0
		if row.JamTidakHadir != nil {
			tdk = *row.JamTidakHadir
		}
		jml := int(math.Round(float64(*row.JamMengajar)*effOrDefault(row.EffectiveWeeks))) - tdk
		resp.JmlHadir = &jml
	}
	return resp
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func toItemResponse(item *PayslipItem) *PayslipItemResponse {
	return &PayslipItemResponse{
		ID: item.ID, ComponentID: item.ComponentID, Code: item.CodeSnapshot,
		Name: item.NameSnapshot, Type: item.Type, Amount: item.Amount, Note: item.Note,
		IsAuto: item.IsAuto,
	}
}

func (s *SlipService) ListByPeriod(ctx context.Context, periodID int64, q string, page, per int64) (*PayslipListResult, error) {
	rows, total, err := s.slip.ListByPeriod(ctx, periodID, q, per, (page-1)*per)
	if err != nil {
		return nil, err
	}
	resp := make([]*PayslipResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toPayslipResponse(row))
	}
	return &PayslipListResult{Payslips: resp, Meta: httpx.NewMeta(page, per, total)}, nil
}

func (s *SlipService) GetSlip(ctx context.Context, slipID int64) (*PayslipDetailResponse, error) {
	row, err := s.slip.FindSlip(ctx, slipID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	items, err := s.slip.ListItems(ctx, slipID)
	if err != nil {
		return nil, err
	}
	periodRow, err := s.repo.FindPeriodRow(ctx, row.PeriodID)
	if err != nil {
		return nil, mapRepoError(err)
	}

	resp := &PayslipDetailResponse{
		PayslipResponse: *toPayslipResponse(row),
		Period: &PeriodBrief{
			ID: periodRow.ID, Name: periodRow.Name, Status: periodRow.Status,
			PeriodMonth: periodRow.PeriodMonth, PeriodYear: periodRow.PeriodYear,
		},
	}
	for _, item := range items {
		resp.Items = append(resp.Items, toItemResponse(item))
	}
	return resp, nil
}

func (s *SlipService) ensureDraft(ctx context.Context, slipID int64) (*payslipRow, error) {
	row, err := s.slip.FindSlip(ctx, slipID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	period, err := s.repo.FindPeriod(ctx, row.PeriodID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	if period.Status != StatusDraft {
		return nil, errors.New("slip hanya bisa diubah pada periode berstatus draft")
	}
	return row, nil
}

func (s *SlipService) UpdateSlipNote(ctx context.Context, slipID int64, note *string) error {
	if _, err := s.ensureDraft(ctx, slipID); err != nil {
		return err
	}
	return s.slip.UpdateSlipNote(ctx, slipID, cleanNote(note))
}

func (s *SlipService) AddItem(ctx context.Context, slipID int64, req SaveItemRequest) (*PayslipDetailResponse, error) {
	if _, err := s.ensureDraft(ctx, slipID); err != nil {
		return nil, err
	}
	if req.Amount <= 0 {
		return nil, errors.New("nominal harus lebih dari 0")
	}

	item := &PayslipItem{PayslipID: slipID, Amount: req.Amount, Note: cleanNote(req.Note)}

	if req.ComponentID != nil && *req.ComponentID > 0 {
		comp, err := s.repo.FindComponent(ctx, *req.ComponentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("komponen tidak ditemukan")
			}
			return nil, err
		}
		componentID := comp.ID
		item.ComponentID = &componentID
		item.CodeSnapshot = comp.Code
		item.NameSnapshot = comp.Name
		item.Type = comp.Type
	} else {
		name := ""
		if req.Name != nil {
			name = *req.Name
		}
		if name == "" {
			return nil, errors.New("nama komponen wajib diisi untuk item manual")
		}
		itemType := ""
		if req.Type != nil {
			itemType = *req.Type
		}
		if !IsValidComponentType(itemType) {
			return nil, errors.New("tipe item harus earning atau deduction")
		}
		item.CodeSnapshot = "MANUAL"
		item.NameSnapshot = name
		item.Type = itemType
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.slip.CreateItem(ctx, item); err != nil {
			return err
		}
		return s.slip.RecomputeTotals(ctx, tx, slipID)
	})
	if err != nil {
		return nil, err
	}
	return s.GetSlip(ctx, slipID)
}

func (s *SlipService) UpdateItem(ctx context.Context, slipID, itemID int64, req UpdateItemRequest) (*PayslipDetailResponse, error) {
	if _, err := s.ensureDraft(ctx, slipID); err != nil {
		return nil, err
	}
	item, err := s.slip.FindItem(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if item.PayslipID != slipID {
		return nil, ErrNotFound
	}
	if req.Amount <= 0 {
		return nil, errors.New("nominal harus lebih dari 0")
	}

	item.Amount = req.Amount
	item.Note = cleanNote(req.Note)
	item.IsAuto = false
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.slip.UpdateItem(ctx, item); err != nil {
			return err
		}
		return s.slip.RecomputeTotals(ctx, tx, slipID)
	})
	if err != nil {
		return nil, err
	}
	return s.GetSlip(ctx, slipID)
}

func (s *SlipService) DeleteItem(ctx context.Context, slipID, itemID int64) (*PayslipDetailResponse, error) {
	if _, err := s.ensureDraft(ctx, slipID); err != nil {
		return nil, err
	}
	item, err := s.slip.FindItem(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if item.PayslipID != slipID {
		return nil, ErrNotFound
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.slip.DeleteItem(ctx, itemID); err != nil {
			return err
		}
		return s.slip.RecomputeTotals(ctx, tx, slipID)
	})
	if err != nil {
		return nil, err
	}
	return s.GetSlip(ctx, slipID)
}

// UpdateAttendance menyimpan data jam (informasional) seluruh slip periode — hanya saat draft.
func (s *SlipService) UpdateAttendance(ctx context.Context, periodID int64, items []AttendanceItem) (int, error) {
	period, err := s.repo.FindPeriod(ctx, periodID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	if period.Status != StatusDraft {
		return 0, errors.New("data jam hanya bisa diubah pada periode berstatus draft")
	}
	if len(items) == 0 {
		return 0, errors.New("tidak ada data jam yang dikirim")
	}
	for _, item := range items {
		if item.PayslipID <= 0 {
			return 0, errors.New("payslip_id tidak valid")
		}
		if neg(item.JamMengajar) || neg(item.JamTidakHadir) || neg(item.HariHadir) {
			return 0, errors.New("nilai jam tidak boleh negatif")
		}
	}

	count := 0
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			ok, err := s.slip.UpdateAttendance(ctx, tx, item.PayslipID, periodID, item.JamMengajar, item.JamTidakHadir, item.HariHadir)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("slip id %d tidak ditemukan pada periode ini", item.PayslipID)
			}
			count++
		}
		return s.syncAssignmentItems(ctx, tx, periodID)
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// syncAssignmentItems menyamakan item slip dengan komponen pegawai terkini:
// membuat item yang belum ada, memperbarui nominal item otomatis (berbasis tarif
// maupun perubahan nominal tetap), dan menghapus item otomatis yang komponennya
// sudah dicabut dari pegawai. Item manual (is_auto=false) tidak pernah disentuh.
func (s *SlipService) syncAssignmentItems(ctx context.Context, tx *gorm.DB, periodID int64) error {
	period, err := s.repo.FindPeriod(ctx, periodID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	workDays, wdErr := s.repo.SchoolWorkingDays(ctx)
	if wdErr != nil {
		return wdErr
	}
	cal := calendarFor(period, workDays)

	slips := []Payslip{}
	if err := tx.WithContext(ctx).Raw(`
		SELECT id, employee_id, jam_mengajar, jam_tidak_hadir, hari_hadir
		FROM payslips WHERE period_id = ?
	`, periodID).Scan(&slips).Error; err != nil {
		return err
	}
	if len(slips) == 0 {
		return nil
	}

	assignments, err := s.slip.ListAssignments(ctx)
	if err != nil {
		return err
	}
	assignmentsByEmployee := map[int64][]*assignmentRow{}
	for _, a := range assignments {
		assignmentsByEmployee[a.EmployeeID] = append(assignmentsByEmployee[a.EmployeeID], a)
	}

	for _, slip := range slips {
		changed := false
		keepIDs := make([]int64, 0, len(assignmentsByEmployee[slip.EmployeeID]))

		for _, a := range assignmentsByEmployee[slip.EmployeeID] {
			keepIDs = append(keepIDs, a.ComponentID)
			amount, note, _ := resolveAssignmentAmount(a, slip.JamMengajar, slip.JamTidakHadir, slip.HariHadir, cal)

			item, err := s.slip.FindAutoItem(ctx, tx, slip.ID, a.ComponentID)
			if err != nil {
				return err
			}
			if item == nil {
				componentID := a.ComponentID
				newItem := &PayslipItem{
					PayslipID:    slip.ID,
					ComponentID:  &componentID,
					CodeSnapshot: a.Code,
					NameSnapshot: a.Name,
					Type:         a.Type,
					Amount:       amount,
					Note:         note,
					IsAuto:       true,
				}
				if err := s.slip.CreateItemTx(ctx, tx, newItem); err != nil {
					return err
				}
				changed = true
				continue
			}
			if item.Amount != amount {
				if err := s.slip.UpdateAutoItemAmount(ctx, tx, item.ID, amount, note); err != nil {
					return err
				}
				changed = true
			}
		}

		deleted, err := s.slip.DeleteOrphanAutoItems(ctx, tx, slip.ID, keepIDs)
		if err != nil {
			return err
		}
		if deleted > 0 {
			changed = true
		}

		if changed {
			if err := s.slip.RecomputeTotals(ctx, tx, slip.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func neg(v *int) bool { return v != nil && *v < 0 }

// Recap membangun rekap per periode: pivot komponen per pegawai + total keseluruhan.
func (s *SlipService) Recap(ctx context.Context, periodID int64) (*RecapResponse, error) {
	periodRow, err := s.repo.FindPeriodRow(ctx, periodID)
	if err != nil {
		return nil, mapRepoError(err)
	}

	slips, _, err := s.slip.ListByPeriod(ctx, periodID, "", 100000, 0)
	if err != nil {
		return nil, err
	}

	recap := &RecapResponse{Period: toPeriodResponse(periodRow)}
	codeIndex := map[string]int{}
	codeOrder := []string{}

	for _, slip := range slips {
		items, err := s.slip.ListItems(ctx, slip.ID)
		if err != nil {
			return nil, err
		}
		row := &RecapRow{
			SlipNo: slip.SlipNo, EmployeeNo: slip.EmployeeNo, EmployeeName: slip.EmployeeName,
			Components: map[string]int64{},
			TotalEarning: slip.TotalEarning, TotalDeduction: slip.TotalDeduction,
			TakeHomePay: slip.TakeHomePay,
		}
		for _, item := range items {
			row.Components[item.CodeSnapshot] += item.Amount
			if _, ok := codeIndex[item.CodeSnapshot]; !ok {
				codeIndex[item.CodeSnapshot] = len(codeOrder)
				codeOrder = append(codeOrder, item.CodeSnapshot)
			}
		}
		recap.Rows = append(recap.Rows, row)
		recap.GrandEarning += slip.TotalEarning
		recap.GrandDeduction += slip.TotalDeduction
		recap.GrandTakeHome += slip.TakeHomePay
	}
	recap.Columns = codeOrder
	return recap, nil
}
