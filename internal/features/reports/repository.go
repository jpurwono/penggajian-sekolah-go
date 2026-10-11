package reports

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type PayrollPeriodRow struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	PeriodMonth    int    `json:"period_month"`
	Status         string `json:"status"`
	SlipCount      int64  `json:"slip_count"`
	TotalEarning   int64  `json:"total_earning"`
	TotalDeduction int64  `json:"total_deduction"`
	TotalTakeHome  int64  `json:"total_take_home"`
}

type PayrollGrand struct {
	SlipCount      int64 `json:"slip_count"`
	TotalEarning   int64 `json:"total_earning"`
	TotalDeduction int64 `json:"total_deduction"`
	TotalTakeHome  int64 `json:"total_take_home"`
}

type PayrollSummary struct {
	Year     int                `json:"year"`
	Periods  []*PayrollPeriodRow `json:"periods"`
	Grand    *PayrollGrand      `json:"grand"`
}

func (r *Repository) PayrollSummary(ctx context.Context, year int) (*PayrollSummary, error) {
	rows := []*PayrollPeriodRow{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.id, p.name, p.period_month, p.status,
		       COUNT(s.id) AS slip_count,
		       COALESCE(SUM(s.total_earning), 0) AS total_earning,
		       COALESCE(SUM(s.total_deduction), 0) AS total_deduction,
		       COALESCE(SUM(s.take_home_pay), 0) AS total_take_home
		FROM payroll_periods p
		LEFT JOIN payslips s ON s.period_id = p.id
		WHERE p.deleted_at IS NULL AND p.period_year = ?
		GROUP BY p.id
		ORDER BY p.period_month ASC
	`, year).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	grand := &PayrollGrand{}
	for _, row := range rows {
		grand.SlipCount += row.SlipCount
		grand.TotalEarning += row.TotalEarning
		grand.TotalDeduction += row.TotalDeduction
		grand.TotalTakeHome += row.TotalTakeHome
	}
	return &PayrollSummary{Year: year, Periods: rows, Grand: grand}, nil
}

type PurchaseMonthRow struct {
	Month     int   `json:"month"`
	Count     int64 `json:"count"`
	TotalCost int64 `json:"total_cost"`
}

type PurchasesSummary struct {
	Year       int                 `json:"year"`
	Months     []*PurchaseMonthRow `json:"months"`
	GrandTotal int64               `json:"grand_total"`
}

func (r *Repository) PurchasesSummary(ctx context.Context, year int) (*PurchasesSummary, error) {
	rows := []*PurchaseMonthRow{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXTRACT(MONTH FROM purchase_date)::int AS month,
		       COUNT(*) AS count,
		       COALESCE(SUM(total_cost), 0) AS total_cost
		FROM inventory_purchases
		WHERE status = 'received' AND EXTRACT(YEAR FROM purchase_date) = ?
		GROUP BY 1
		ORDER BY 1
	`, year).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	summary := &PurchasesSummary{Year: year, Months: rows}
	for _, row := range rows {
		summary.GrandTotal += row.TotalCost
	}
	return summary, nil
}

type EmployeeTypeRow struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

func (r *Repository) EmployeeTypeCounts(ctx context.Context) ([]*EmployeeTypeRow, error) {
	rows := []*EmployeeTypeRow{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT employee_type AS type, COUNT(*) AS count
		FROM employees
		WHERE deleted_at IS NULL AND is_active = true
		GROUP BY employee_type
		ORDER BY count DESC
	`).Scan(&rows).Error
	return rows, err
}

type BrokenAssetRow struct {
	ItemCode     string `json:"item_code"`
	ItemName     string `json:"item_name"`
	Unit         string `json:"unit"`
	LocationName string `json:"location_name"`
	Condition    string `json:"condition"`
	Quantity     int    `json:"quantity"`
}

func (r *Repository) BrokenAssets(ctx context.Context, condition string) ([]*BrokenAssetRow, error) {
	where := "AND 1=1"
	args := []any{}
	if condition == "rusak_ringan" || condition == "rusak_berat" {
		where = "AND s.condition = ?"
		args = append(args, condition)
	}

	rows := []*BrokenAssetRow{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT i.code AS item_code, i.name AS item_name, i.unit,
		       l.name AS location_name, s.condition, s.quantity
		FROM inventory_stocks s
		JOIN inventory_items i ON i.id = s.item_id AND i.deleted_at IS NULL
		JOIN inventory_locations l ON l.id = s.location_id
		WHERE s.condition IN ('rusak_ringan', 'rusak_berat') AND s.quantity > 0
		`+where+`
		ORDER BY i.name ASC, l.name ASC
	`, args...).Scan(&rows).Error
	return rows, err
}
