package payroll

import "testing"

func TestFirstWorkdaysEnd(t *testing.T) {
	cases := []struct {
		year, month int
		wantDay     int // hari setelah hari kerja ke-5 (batas eksklusif), Jakarta
	}{
		{2026, 10, 8},  // 1=Kamis → 1,2,5,6,7 → batas tgl 8
		{2026, 11, 7},  // 1=Minggu → 2,3,4,5,6 → batas tgl 7
		{2026, 9, 8},   // 1=Selasa → 1,2,3,4,7 → batas tgl 8
		{2026, 8, 8},   // 1=Sabtu → 3,4,5,6,7 → batas tgl 8
	}
	for _, c := range cases {
		got, err := FirstWorkdaysEnd(c.year, c.month, 5)
		if err != nil {
			t.Fatal(err)
		}
		if got.Day() != c.wantDay {
			t.Errorf("%d-%02d: batas = tgl %d, mau %d", c.year, c.month, got.Day(), c.wantDay)
		}
	}
}
