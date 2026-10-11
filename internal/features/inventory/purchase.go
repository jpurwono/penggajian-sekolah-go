package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"penggajian/server/internal/platform/audit"
	"penggajian/server/internal/platform/middleware"
)

type PurchaseService struct {
	db   *gorm.DB
	repo *Repository
}

func NewPurchaseService(db *gorm.DB, repo *Repository) *PurchaseService {
	return &PurchaseService{db: db, repo: repo}
}

const purchaseSelect = `
SELECT p.*, (SELECT COUNT(*) FROM inventory_purchase_items pi WHERE pi.purchase_id = p.id) AS item_count
FROM inventory_purchases p `

type purchaseListRow struct {
	Purchase
	ItemCount int `gorm:"column:item_count"`
}

func nextPurchaseNo(tx *gorm.DB, date time.Time) (string, error) {
	prefix := fmt.Sprintf("PRC-%d%02d-", date.Year(), int(date.Month()))
	var count int64
	if err := tx.Raw(`SELECT COUNT(*) FROM inventory_purchases WHERE purchase_no LIKE ?`, prefix+"%").Scan(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (s *PurchaseService) validateItems(ctx context.Context, items []PurchaseItemInput) ([]*PurchaseItem, error) {
	if len(items) == 0 {
		return nil, errors.New("minimal satu baris barang wajib diisi")
	}
	seen := map[int64]bool{}
	result := make([]*PurchaseItem, 0, len(items))
	for _, input := range items {
		if input.ItemID <= 0 || input.Quantity <= 0 || input.LocationID <= 0 {
			return nil, errors.New("setiap baris wajib: barang, jumlah > 0, dan lokasi tujuan")
		}
		if input.UnitPrice < 0 {
			return nil, errors.New("harga satuan tidak boleh negatif")
		}
		if seen[input.ItemID] {
			return nil, errors.New("ada barang yang terinput dua kali, gabungkan dalam satu baris")
		}
		seen[input.ItemID] = true

		if _, err := s.repo.FindItem(ctx, input.ItemID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("barang tidak ditemukan pada salah satu baris")
			}
			return nil, err
		}
		if _, err := s.repo.FindLocation(ctx, input.LocationID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("lokasi tujuan tidak ditemukan pada salah satu baris")
			}
			return nil, err
		}

		result = append(result, &PurchaseItem{
			ItemID:     input.ItemID,
			Quantity:   input.Quantity,
			UnitPrice:  input.UnitPrice,
			Subtotal:   input.UnitPrice * int64(input.Quantity),
			LocationID: input.LocationID,
		})
	}
	return result, nil
}

func (s *PurchaseService) List(ctx context.Context, q, status string, limit, offset int64) ([]*purchaseListRow, int64, error) {
	where := "WHERE 1=1"
	args := []any{}
	if q != "" {
		where += " AND (p.purchase_no ILIKE ? OR COALESCE(p.supplier_name, '') ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	if status != "" {
		where += " AND p.status = ?"
		args = append(args, status)
	}

	var total int64
	if err := s.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM inventory_purchases p "+where, args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*purchaseListRow{}
	err := s.db.WithContext(ctx).Raw(
		purchaseSelect+where+" ORDER BY p.id DESC LIMIT ? OFFSET ?",
		append(args, limit, offset)...,
	).Scan(&rows).Error
	return rows, total, err
}

type purchaseItemRow struct {
	PurchaseItem
	ItemCode     string  `gorm:"column:item_code"`
	ItemName     string  `gorm:"column:item_name"`
	Unit         string  `gorm:"column:unit"`
	LocationName *string `gorm:"column:location_name"`
}

func (s *PurchaseService) Find(ctx context.Context, id int64) (*Purchase, []*purchaseItemRow, error) {
	var p Purchase
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, nil, err
	}
	items := []*purchaseItemRow{}
	err := s.db.WithContext(ctx).Raw(`
		SELECT pi.*, i.code AS item_code, i.name AS item_name, i.unit, l.name AS location_name
		FROM inventory_purchase_items pi
		JOIN inventory_items i ON i.id = pi.item_id
		LEFT JOIN inventory_locations l ON l.id = pi.location_id
		WHERE pi.purchase_id = ? ORDER BY pi.id ASC
	`, id).Scan(&items).Error
	return &p, items, err
}

func (s *PurchaseService) Create(ctx context.Context, c echo.Context, req SavePurchaseRequest) (int64, error) {
	purchaseDate := time.Now()
	if req.PurchaseDate != "" {
		var err error
		purchaseDate, err = time.Parse(dateLayout, req.PurchaseDate)
		if err != nil {
			return 0, errors.New("format tanggal harus YYYY-MM-DD")
		}
	}

	items, err := s.validateItems(ctx, req.Items)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, item := range items {
		total += item.Subtotal
	}

	purchase := &Purchase{
		PurchaseDate: purchaseDate,
		SupplierName: cleanStr(&req.SupplierName),
		Description:  cleanStr(&req.Description),
		TotalCost:    total,
		Status:       PurchaseDraft,
	}
	actorID := middleware.CurrentUserID(c)
	purchase.CreatedBy = &actorID

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		purchaseNo, err := nextPurchaseNo(tx, purchaseDate)
		if err != nil {
			return err
		}
		purchase.PurchaseNo = purchaseNo

		if err := tx.Create(purchase).Error; err != nil {
			return err
		}
		for _, item := range items {
			item.PurchaseID = purchase.ID
		}
		return tx.Create(&items).Error
	})
	if err != nil {
		return 0, err
	}

	audit.Log(c, s.db, "create", "inventory_purchases", purchase.ID, nil, purchase)
	return purchase.ID, nil
}

func (s *PurchaseService) Update(ctx context.Context, c echo.Context, id int64, req SavePurchaseRequest) error {
	existing, _, err := s.Find(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if existing.Status != PurchaseDraft {
		return errors.New("hanya pengadaan berstatus draft yang dapat diubah")
	}

	purchaseDate := existing.PurchaseDate
	if req.PurchaseDate != "" {
		purchaseDate, err = time.Parse(dateLayout, req.PurchaseDate)
		if err != nil {
			return errors.New("format tanggal harus YYYY-MM-DD")
		}
	}

	items, err := s.validateItems(ctx, req.Items)
	if err != nil {
		return err
	}

	var total int64
	for _, item := range items {
		total += item.Subtotal
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Purchase{}).Where("id = ?", id).Updates(map[string]any{
			"purchase_date": purchaseDate,
			"supplier_name": cleanStr(&req.SupplierName),
			"description":   cleanStr(&req.Description),
			"total_cost":    total,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("purchase_id = ?", id).Delete(&PurchaseItem{}).Error; err != nil {
			return err
		}
		for _, item := range items {
			item.PurchaseID = id
		}
		return tx.Create(&items).Error
	})
}

func (s *PurchaseService) Delete(ctx context.Context, id int64) error {
	existing, _, err := s.Find(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if existing.Status != PurchaseDraft {
		return errors.New("hanya pengadaan berstatus draft yang dapat dihapus")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("purchase_id = ?", id).Delete(&PurchaseItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&Purchase{}, id).Error
	})
}

// Receive menandai pengadaan diterima: tiap baris menjadi mutasi masuk + stok bertambah.
func (s *PurchaseService) Receive(ctx context.Context, c echo.Context, id int64) error {
	purchase, items, err := s.Find(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if purchase.Status != PurchaseDraft {
		return errors.New("pengadaan sudah pernah diterima")
	}

	actorID := middleware.CurrentUserID(c)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			mutationNo, err := nextMutationNo(tx, purchase.PurchaseDate)
			if err != nil {
				return err
			}
			note := fmt.Sprintf("Penerimaan pengadaan %s", purchase.PurchaseNo)
			mutation := &Mutation{
				MutationNo:   mutationNo,
				Type:         MutIn,
				MutationDate: purchase.PurchaseDate,
				ItemID:       item.ItemID,
				ToLocationID: &item.LocationID,
				Quantity:     item.Quantity,
				ReferenceNo:  &purchase.PurchaseNo,
				Note:         &note,
				CreatedBy:    &actorID,
			}
			if err := tx.Create(mutation).Error; err != nil {
				return err
			}
			if err := s.increaseStockTx(tx, item.ItemID, stockDelta{
				LocationID: item.LocationID, Condition: CondBaik, Delta: item.Quantity,
			}); err != nil {
				return err
			}
		}
		if err := tx.Exec(`UPDATE inventory_purchases SET status = ? WHERE id = ?`, PurchaseReceived, id).Error; err != nil {
			return err
		}
		audit.Log(c, s.db, "receive", "inventory_purchases", id, nil, map[string]any{
			"purchase_no": purchase.PurchaseNo, "items": len(items),
		})
		return nil
	})
}

func (s *PurchaseService) increaseStockTx(tx *gorm.DB, itemID int64, d stockDelta) error {
	return tx.Exec(`
		INSERT INTO inventory_stocks (item_id, location_id, condition, quantity)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (item_id, location_id, condition)
		DO UPDATE SET quantity = inventory_stocks.quantity + EXCLUDED.quantity
	`, itemID, d.LocationID, d.Condition, d.Delta).Error
}

func toPurchaseResponse(p *purchaseListRow) *PurchaseResponse {
	return &PurchaseResponse{
		ID: p.ID, PurchaseNo: p.PurchaseNo,
		PurchaseDate: p.PurchaseDate.Format(dateLayout),
		SupplierName: p.SupplierName, Description: p.Description,
		TotalCost: p.TotalCost, Status: p.Status, ItemCount: p.ItemCount,
		CreatedAt: p.CreatedAt,
	}
}

func toPurchaseDetail(p *Purchase, items []*purchaseItemRow) *PurchaseDetailResponse {
	resp := &PurchaseDetailResponse{
		PurchaseResponse: PurchaseResponse{
			ID: p.ID, PurchaseNo: p.PurchaseNo,
			PurchaseDate: p.PurchaseDate.Format(dateLayout),
			SupplierName: p.SupplierName, Description: p.Description,
			TotalCost: p.TotalCost, Status: p.Status, ItemCount: len(items),
			CreatedAt: p.CreatedAt,
		},
		Items: make([]*PurchaseItemResponse, 0, len(items)),
	}
	for _, item := range items {
		resp.Items = append(resp.Items, &PurchaseItemResponse{
			ID: item.ID, ItemID: item.ItemID, ItemCode: item.ItemCode, ItemName: item.ItemName,
			Unit: item.Unit, Quantity: item.Quantity, UnitPrice: item.UnitPrice,
			Subtotal: item.Subtotal, LocationID: item.LocationID, LocationName: derefStr(item.LocationName),
		})
	}
	return resp
}
