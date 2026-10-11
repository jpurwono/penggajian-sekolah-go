package inventory

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func cleanStr(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---------- Kategori ----------

type categoryRow struct {
	Category
	ParentName *string `gorm:"column:parent_name"`
}

func (r *Repository) ListCategories(ctx context.Context, q string, limit, offset int64) ([]*categoryRow, int64, error) {
	where := "WHERE c.deleted_at IS NULL"
	args := []any{}
	if q != "" {
		where += " AND (c.name ILIKE ? OR COALESCE(c.code, '') ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like)
	}

	var total int64
	if err := r.dbCtx(ctx).Raw(
		"SELECT COUNT(*) FROM inventory_categories c "+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*categoryRow{}
	err := r.dbCtx(ctx).Raw(`
		SELECT c.*, p.name AS parent_name
		FROM inventory_categories c
		LEFT JOIN inventory_categories p ON p.id = c.parent_id
		`+where+` ORDER BY c.name ASC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) FindCategory(ctx context.Context, id int64) (*Category, error) {
	var c Category
	if err := r.dbCtx(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) CreateCategory(ctx context.Context, c *Category) error {
	return r.dbCtx(ctx).Create(c).Error
}

func (r *Repository) UpdateCategory(ctx context.Context, c *Category) error {
	return r.dbCtx(ctx).Model(c).
		Select("code", "name", "parent_id", "description").Updates(c).Error
}

func (r *Repository) SoftDeleteCategory(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&Category{}, id).Error
}

func (r *Repository) CountItemsInCategory(ctx context.Context, categoryID int64) (int64, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&Item{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

func (r *Repository) CountChildCategories(ctx context.Context, parentID int64) (int64, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&Category{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// ---------- Lokasi ----------

func (r *Repository) ListLocations(ctx context.Context, q string, limit, offset int64) ([]*Location, int64, error) {
	query := r.dbCtx(ctx).Model(&Location{})
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("name ILIKE ? OR COALESCE(code, '') ILIKE ? OR COALESCE(building, '') ILIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*Location
	err := query.Order("name ASC").Offset(int(offset)).Limit(int(limit)).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindLocation(ctx context.Context, id int64) (*Location, error) {
	var l Location
	if err := r.dbCtx(ctx).First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *Repository) CreateLocation(ctx context.Context, l *Location) error {
	return r.dbCtx(ctx).Create(l).Error
}

func (r *Repository) UpdateLocation(ctx context.Context, l *Location) error {
	return r.dbCtx(ctx).Model(l).
		Select("code", "name", "building", "floor", "description").Updates(l).Error
}

func (r *Repository) SoftDeleteLocation(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&Location{}, id).Error
}

func (r *Repository) LocationHasStock(ctx context.Context, locationID int64) (bool, error) {
	var total int
	err := r.dbCtx(ctx).Raw(
		"SELECT COALESCE(SUM(quantity), 0) FROM inventory_stocks WHERE location_id = ?", locationID,
	).Scan(&total).Error
	return total > 0, err
}

func (r *Repository) LocationReferenced(ctx context.Context, locationID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Raw(`
		SELECT (SELECT COUNT(*) FROM inventory_mutations m
		        WHERE m.from_location_id = ? OR m.to_location_id = ?)
		     + (SELECT COUNT(*) FROM inventory_purchase_items pi WHERE pi.location_id = ?)
	`, locationID, locationID, locationID).Scan(&count).Error
	return count > 0, err
}

// ---------- Item ----------

type itemRow struct {
	Item
	CategoryName *string `gorm:"column:category_name"`
	TotalQty     int     `gorm:"column:total_qty"`
}

const itemSelect = `
SELECT i.*, c.name AS category_name,
       COALESCE((SELECT SUM(s.quantity) FROM inventory_stocks s WHERE s.item_id = i.id), 0) AS total_qty
FROM inventory_items i
LEFT JOIN inventory_categories c ON c.id = i.category_id `

func (r *Repository) ListItems(ctx context.Context, q string, categoryID *int64, active *bool, limit, offset int64) ([]*itemRow, int64, error) {
	where := "WHERE i.deleted_at IS NULL"
	args := []any{}
	if q != "" {
		where += " AND (i.name ILIKE ? OR i.code ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	if categoryID != nil {
		where += " AND i.category_id = ?"
		args = append(args, *categoryID)
	}
	if active != nil {
		where += " AND i.is_active = ?"
		args = append(args, *active)
	}

	var total int64
	if err := r.dbCtx(ctx).Raw(
		"SELECT COUNT(*) FROM inventory_items i "+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*itemRow{}
	err := r.dbCtx(ctx).Raw(
		itemSelect+where+" ORDER BY i.name ASC LIMIT ? OFFSET ?",
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) FindItem(ctx context.Context, id int64) (*Item, error) {
	var i Item
	if err := r.dbCtx(ctx).First(&i, id).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *Repository) CreateItem(ctx context.Context, i *Item) error {
	return r.dbCtx(ctx).Create(i).Error
}

func (r *Repository) UpdateItem(ctx context.Context, i *Item) error {
	return r.dbCtx(ctx).Model(i).
		Select("code", "name", "category_id", "unit", "min_stock", "description", "is_active").Updates(i).Error
}

func (r *Repository) SoftDeleteItem(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&Item{}, id).Error
}

func (r *Repository) ItemCodeExists(ctx context.Context, code string, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Unscoped().Model(&Item{}).
		Where("code = ? AND id <> ?", code, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) ItemTotalStock(ctx context.Context, itemID int64) (int, error) {
	var total int
	err := r.dbCtx(ctx).Raw(
		"SELECT COALESCE(SUM(quantity), 0) FROM inventory_stocks WHERE item_id = ?", itemID,
	).Scan(&total).Error
	return total, err
}

// ---------- Stok ----------

func (r *Repository) ListStocks(ctx context.Context, q string, itemID, locationID *int64, condition string, limit, offset int64) ([]*StockResponse, int64, error) {
	where := "WHERE s.quantity > 0"
	args := []any{}

	if q != "" {
		where += " AND (i.name ILIKE ? OR i.code ILIKE ? OR l.name ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	if itemID != nil {
		where += " AND s.item_id = ?"
		args = append(args, *itemID)
	}
	if locationID != nil {
		where += " AND s.location_id = ?"
		args = append(args, *locationID)
	}
	if condition != "" {
		where += " AND s.condition = ?"
		args = append(args, condition)
	}

	var total int64
	if err := r.dbCtx(ctx).Raw(`
		SELECT COUNT(*) FROM inventory_stocks s
		JOIN inventory_items i ON i.id = s.item_id AND i.deleted_at IS NULL
		JOIN inventory_locations l ON l.id = s.location_id `+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*StockResponse{}
	err := r.dbCtx(ctx).Raw(`
		SELECT s.id, s.item_id, i.code AS item_code, i.name AS item_name, i.unit,
		       s.location_id, l.name AS location_name, s.condition, s.quantity
		FROM inventory_stocks s
		JOIN inventory_items i ON i.id = s.item_id AND i.deleted_at IS NULL
		JOIN inventory_locations l ON l.id = s.location_id
		`+where+` ORDER BY i.name ASC, l.name ASC, s.condition ASC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) StockSummary(ctx context.Context, q string, belowMin bool, limit, offset int64) ([]*StockSummaryRow, int64, error) {
	where := `WHERE i.deleted_at IS NULL AND i.is_active = true`
	args := []any{}
	if q != "" {
		where += " AND (i.name ILIKE ? OR i.code ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like)
	}

	having := ""
	if belowMin {
		having = ` HAVING COALESCE(SUM(s.quantity), 0) < i.min_stock AND i.min_stock > 0`
	}

	baseQuery := `
		FROM inventory_items i
		LEFT JOIN inventory_stocks s ON s.item_id = i.id
		` + where + `
		GROUP BY i.id` + having

	var total int64
	if err := r.dbCtx(ctx).Raw("SELECT COUNT(*) FROM (SELECT i.id"+baseQuery+") t", args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*StockSummaryRow{}
	err := r.dbCtx(ctx).Raw(`
		SELECT i.id, i.code, i.name, i.unit, i.min_stock, i.is_active, i.created_at,
		       COALESCE(SUM(s.quantity), 0) AS total_qty,
		       COALESCE(SUM(s.quantity) FILTER (WHERE s.condition = 'baik'), 0) AS qty_baik,
		       COALESCE(SUM(s.quantity) FILTER (WHERE s.condition = 'rusak_ringan'), 0) AS qty_ringan,
		       COALESCE(SUM(s.quantity) FILTER (WHERE s.condition = 'rusak_berat'), 0) AS qty_berat
		`+baseQuery+` ORDER BY i.name ASC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...,
	).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	for _, row := range rows {
		row.BelowMin = row.MinStock > 0 && row.TotalQty < row.MinStock
	}
	return rows, total, nil
}

// ---------- Daftar Mutasi ----------

func (r *Repository) ListMutations(ctx context.Context, q string, itemType string, itemID *int64, dateFrom, dateTo string, limit, offset int64) ([]*MutationResponse, int64, error) {
	where := "WHERE 1=1"
	args := []any{}

	if q != "" {
		where += " AND (m.mutation_no ILIKE ? OR COALESCE(m.reference_no, '') ILIKE ? OR i.name ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	if itemType != "" {
		where += " AND m.type = ?"
		args = append(args, itemType)
	}
	if itemID != nil {
		where += " AND m.item_id = ?"
		args = append(args, *itemID)
	}
	if dateFrom != "" {
		where += " AND m.mutation_date >= ?"
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		where += " AND m.mutation_date <= ?"
		args = append(args, dateTo)
	}

	joinedFrom := `
		FROM inventory_mutations m
		JOIN inventory_items i ON i.id = m.item_id
		LEFT JOIN inventory_locations fl ON fl.id = m.from_location_id
		LEFT JOIN inventory_locations tl ON tl.id = m.to_location_id `

	var total int64
	if err := r.dbCtx(ctx).Raw("SELECT COUNT(*) "+joinedFrom+where, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*MutationResponse{}
	err := r.dbCtx(ctx).Raw(`
		SELECT m.id, m.mutation_no, m.type, m.mutation_date, m.item_id,
		       i.code AS item_code, i.name AS item_name,
		       m.from_location_id, fl.name AS from_location,
		       m.to_location_id, tl.name AS to_location,
		       m.from_condition, m.to_condition,
		       m.quantity, m.reference_no, m.note, m.created_at
		`+joinedFrom+where+` ORDER BY m.id DESC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...,
	).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
