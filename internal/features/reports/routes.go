package reports

import (
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
	mw "penggajian/server/internal/platform/middleware"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) PayrollSummary(c echo.Context) error {
	year := int(parseYear(c, time.Now().Year()))
	data, err := h.repo.PayrollSummary(c.Request().Context(), year)
	if err != nil {
		return httpx.ErrInternal(c, "Gagal memuat rekap penggajian")
	}
	return httpx.OK(c, data)
}

func (h *Handler) PurchasesSummary(c echo.Context) error {
	year := int(parseYear(c, time.Now().Year()))
	data, err := h.repo.PurchasesSummary(c.Request().Context(), year)
	if err != nil {
		return httpx.ErrInternal(c, "Gagal memuat rekap pengadaan")
	}
	return httpx.OK(c, data)
}

func (h *Handler) BrokenAssets(c echo.Context) error {
	data, err := h.repo.BrokenAssets(c.Request().Context(), c.QueryParam("condition"))
	if err != nil {
		return httpx.ErrInternal(c, "Gagal memuat laporan barang rusak")
	}
	return httpx.OK(c, data)
}

func (h *Handler) EmployeeTypeCounts(c echo.Context) error {
	data, err := h.repo.EmployeeTypeCounts(c.Request().Context())
	if err != nil {
		return httpx.ErrInternal(c, "Gagal memuat komposisi pegawai")
	}
	return httpx.OK(c, data)
}

func parseYear(c echo.Context, fallback int) int {
	raw := c.QueryParam("year")
	if raw == "" {
		return fallback
	}
	var year int64
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return fallback
		}
		year = year*10 + int64(ch-'0')
	}
	if year < 1900 || year > 2100 {
		return fallback
	}
	return int(year)
}

// Register mendaftarkan routes fitur reports ke group /api/v1.
func Register(g *echo.Group, db *gorm.DB, authMW echo.MiddlewareFunc) {
	h := NewHandler(NewRepository(db))

	grp := g.Group("/reports", authMW, mw.RequirePermission("laporan.read"))
	grp.GET("/payroll-summary", h.PayrollSummary)
	grp.GET("/purchases-summary", h.PurchasesSummary)
	grp.GET("/broken-assets", h.BrokenAssets)
	grp.GET("/spp-summary", h.SppSummary)
	grp.GET("/employee-type-counts", h.EmployeeTypeCounts)
}

func (h *Handler) SppSummary(c echo.Context) error {
	year := parseYear(c, time.Now().Year())

	months, err := h.repo.SppMonthlySummary(c.Request().Context(), year)
	if err != nil {
		return httpx.ErrInternal(c, "Gagal memuat rekap bulanan SPP")
	}

	students, err := h.repo.SppStudentSummary(c.Request().Context(), year)
	if err != nil {
		return httpx.ErrInternal(c, "Gagal memuat rekap siswa SPP")
	}

	totals := &SppMonthRow{MonthName: "TOTAL"}
	for _, m := range months {
		totals.TotalBills += m.TotalBills
		totals.PaidCount += m.PaidCount
		totals.PartialCount += m.PartialCount
		totals.UnpaidCount += m.UnpaidCount
		totals.TotalBilled += m.TotalBilled
		totals.TotalCollected += m.TotalCollected
		totals.TotalOutstanding += m.TotalOutstanding
	}

	return httpx.OK(c, &SppSummary{
		Year:     year,
		Months:   months,
		Students: students,
		Totals:   totals,
	})
}
