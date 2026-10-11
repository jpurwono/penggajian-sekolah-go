package payroll

import "time"

// CountWorkingDays menghitung hari kerja sebulan (Senin–Jumat untuk 5 hari,
// Senin–Sabtu untuk 6 hari). Minggu parsial di awal/akhir bulan dihitung
// per tanggal sesuai kalender nyata.
func CountWorkingDays(year, month, workDays int) int {
	if month < 1 || month > 12 {
		return 0
	}
	if workDays != 6 {
		workDays = 5
	}
	lastDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).
		AddDate(0, 1, -1).Day()

	count := 0
	for d := 1; d <= lastDay; d++ {
		wd := int(time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC).Weekday())
		if wd >= 1 && wd <= workDays {
			count++
		}
	}
	return count
}

// EffectiveWeeks menghasilkan jumlah minggu efektif: hari kerja ÷ hari kerja per minggu.
// Contoh: September 2026, 5 hari/minggu → 22 hari → 4,4 minggu.
func EffectiveWeeks(year, month, workDays int) float64 {
	if workDays != 6 {
		workDays = 5
	}
	return float64(CountWorkingDays(year, month, workDays)) / float64(workDays)
}

// FirstWeekWorkingDays menghitung hari kerja pada pekan kalender pertama bulan
// (tanggal 1 s/d Sabtu pertama, dibatasi akhir bulan). Contoh: Senin tgl 1,
// 5 hari/minggu → tgl 1–5 → 5 hari kerja.
func FirstWeekWorkingDays(year, month, workDays int) int {
	if month < 1 || month > 12 {
		return 0
	}
	if workDays != 6 {
		workDays = 5
	}
	lastDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).
		AddDate(0, 1, -1).Day()

	saturday := lastDay
	for d := 1; d <= 7 && d <= lastDay; d++ {
		if time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC).Weekday() == time.Saturday {
			saturday = d
			break
		}
	}

	count := 0
	for d := 1; d <= saturday; d++ {
		wd := int(time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC).Weekday())
		if wd >= 1 && wd <= workDays {
			count++
		}
	}
	return count
}

func effOrDefault(effectiveWeeks *float64) float64 {
	if effectiveWeeks == nil || *effectiveWeeks <= 0 {
		return 4.0
	}
	return *effectiveWeeks
}

// periodCalendar menyatukan parameter kalender yang dipakai perhitungan komponen.
type periodCalendar struct {
	EffectiveWeeks float64
	WorkDays       int
	FirstWeekDays  int
}

func calendarFor(period *PayrollPeriod, workDays int) periodCalendar {
	if workDays != 5 && workDays != 6 {
		workDays = 5
	}
	return periodCalendar{
		EffectiveWeeks: effOrDefault(period.EffectiveWeeks),
		WorkDays:       workDays,
		FirstWeekDays:  FirstWeekWorkingDays(period.PeriodYear, period.PeriodMonth, workDays),
	}
}
