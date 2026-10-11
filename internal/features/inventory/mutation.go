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

const dateLayout = "2006-01-02"

var ErrNotFound = errors.New("data tidak ditemukan")

type MutationService struct {
	db   *gorm.DB
	repo *Repository
}

func NewMutationService(db *gorm.DB, repo *Repository) *MutationService {
	return &MutationService{db: db, repo: repo}
}

// stockDelta menggambarkan perubahan stok pada kombinasi lokasi + kondisi.
type stockDelta struct {
	LocationID int64
	Condition  string
	Delta      int
}

func (s *MutationService) buildDeltas(req CreateMutationRequest) ([]stockDelta, error) {
	baik := CondBaik

	switch req.Type {
	case MutIn:
		if req.ToLocationID == nil {
			return nil, errors.New("lokasi tujuan wajib dipilih untuk mutasi masuk")
		}
		cond := baik
		if req.ToCondition != nil && *req.ToCondition != "" {
			cond = *req.ToCondition
		}
		return []stockDelta{{LocationID: *req.ToLocationID, Condition: cond, Delta: req.Quantity}}, nil

	case MutOut:
		if req.FromLocationID == nil {
			return nil, errors.New("lokasi asal wajib dipilih untuk mutasi keluar")
		}
		cond := baik
		if req.FromCondition != nil && *req.FromCondition != "" {
			cond = *req.FromCondition
		}
		return []stockDelta{{LocationID: *req.FromLocationID, Condition: cond, Delta: -req.Quantity}}, nil

	case MutTransfer:
		if req.FromLocationID == nil || req.ToLocationID == nil {
			return nil, errors.New("lokasi asal & tujuan wajib dipilih untuk mutasi ruangan")
		}
		if *req.FromLocationID == *req.ToLocationID {
			return nil, errors.New("lokasi asal dan tujuan tidak boleh sama")
		}
		cond := baik
		if req.FromCondition != nil && *req.FromCondition != "" {
			cond = *req.FromCondition
		}
		return []stockDelta{
			{LocationID: *req.FromLocationID, Condition: cond, Delta: -req.Quantity},
			{LocationID: *req.ToLocationID, Condition: cond, Delta: req.Quantity},
		}, nil

	case MutRepair:
		if req.FromLocationID == nil {
			return nil, errors.New("lokasi wajib dipilih untuk perbaikan (ubah kondisi)")
		}
		if req.FromCondition == nil || req.ToCondition == nil || *req.FromCondition == "" || *req.ToCondition == "" {
			return nil, errors.New("kondisi awal & kondisi baru wajib dipilih untuk perbaikan")
		}
		if *req.FromCondition == *req.ToCondition {
			return nil, errors.New("kondisi awal dan kondisi baru tidak boleh sama")
		}
		return []stockDelta{
			{LocationID: *req.FromLocationID, Condition: *req.FromCondition, Delta: -req.Quantity},
			{LocationID: *req.FromLocationID, Condition: *req.ToCondition, Delta: req.Quantity},
		}, nil

	case MutDiscard:
		if req.FromLocationID == nil {
			return nil, errors.New("lokasi asal wajib dipilih untuk penghapusan")
		}
		cond := baik
		if req.FromCondition != nil && *req.FromCondition != "" {
			cond = *req.FromCondition
		}
		return []stockDelta{{LocationID: *req.FromLocationID, Condition: cond, Delta: -req.Quantity}}, nil
	}

	return nil, errors.New("tipe mutasi tidak valid")
}

func (s *MutationService) increaseStock(tx *gorm.DB, itemID int64, d stockDelta) error {
	return tx.Exec(`
		INSERT INTO inventory_stocks (item_id, location_id, condition, quantity)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (item_id, location_id, condition)
		DO UPDATE SET quantity = inventory_stocks.quantity + EXCLUDED.quantity
	`, itemID, d.LocationID, d.Condition, d.Delta).Error
}

func (s *MutationService) decreaseStock(tx *gorm.DB, itemID int64, d stockDelta) error {
	res := tx.Exec(`
		UPDATE inventory_stocks SET quantity = quantity - ?
		WHERE item_id = ? AND location_id = ? AND condition = ? AND quantity >= ?
	`, -d.Delta, itemID, d.LocationID, d.Condition, -d.Delta)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var available int
		tx.Raw(`SELECT quantity FROM inventory_stocks WHERE item_id = ? AND location_id = ? AND condition = ?`,
			itemID, d.LocationID, d.Condition).Scan(&available)
		return fmt.Errorf("stok tidak cukup: tersedia %d di lokasi/kondisi tersebut", available)
	}
	return nil
}

func nextMutationNo(tx *gorm.DB, date time.Time) (string, error) {
	prefix := fmt.Sprintf("MUT-%d%02d-", date.Year(), int(date.Month()))
	var count int64
	if err := tx.Raw(`SELECT COUNT(*) FROM inventory_mutations WHERE mutation_no LIKE ?`, prefix+"%").Scan(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (s *MutationService) Create(ctx context.Context, c echo.Context, req CreateMutationRequest) (*MutationResponse, error) {
	if !IsValidMutationType(req.Type) {
		return nil, errors.New("tipe mutasi tidak valid")
	}
	if req.ItemID <= 0 {
		return nil, errors.New("barang wajib dipilih")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("jumlah harus lebih dari 0")
	}

	item, err := s.repo.FindItem(ctx, req.ItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("barang tidak ditemukan")
		}
		return nil, err
	}

	if req.FromCondition != nil && !IsValidCondition(*req.FromCondition) {
		return nil, errors.New("kondisi awal tidak valid")
	}
	if req.ToCondition != nil && !IsValidCondition(*req.ToCondition) {
		return nil, errors.New("kondisi baru tidak valid")
	}

	deltas, err := s.buildDeltas(req)
	if err != nil {
		return nil, err
	}

	locationIDs := map[int64]bool{}
	if req.FromLocationID != nil {
		locationIDs[*req.FromLocationID] = true
	}
	if req.ToLocationID != nil {
		locationIDs[*req.ToLocationID] = true
	}
	for id := range locationIDs {
		if _, err := s.repo.FindLocation(ctx, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("lokasi tidak ditemukan")
			}
			return nil, err
		}
	}

	mutationDate := time.Now()
	if req.MutationDate != "" {
		mutationDate, err = time.Parse(dateLayout, req.MutationDate)
		if err != nil {
			return nil, errors.New("format tanggal harus YYYY-MM-DD")
		}
	}

	mutation := &Mutation{
		Type: req.Type, MutationDate: mutationDate, ItemID: req.ItemID,
		FromLocationID: req.FromLocationID, ToLocationID: req.ToLocationID,
		FromCondition: cleanStr(req.FromCondition), ToCondition: cleanStr(req.ToCondition),
		Quantity: req.Quantity, ReferenceNo: cleanStr(req.ReferenceNo), Note: cleanStr(req.Note),
	}
	actorID := middleware.CurrentUserID(c)
	mutation.CreatedBy = &actorID

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		mutationNo, err := nextMutationNo(tx, mutationDate)
		if err != nil {
			return err
		}
		mutation.MutationNo = mutationNo

		if err := tx.Create(mutation).Error; err != nil {
			return err
		}
		for _, d := range deltas {
			if d.Delta > 0 {
				if err := s.increaseStock(tx, req.ItemID, d); err != nil {
					return err
				}
			} else {
				if err := s.decreaseStock(tx, req.ItemID, d); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	audit.Log(c, s.db, "create", "inventory_mutations", mutation.ID, nil, map[string]any{
		"mutation_no": mutation.MutationNo, "type": mutation.Type, "item": item.Name,
		"quantity": mutation.Quantity,
	})

	list, _, err := s.repo.ListMutations(ctx, "", "", &req.ItemID, "", "", 1, 0)
	if err != nil || len(list) == 0 {
		return &MutationResponse{
			ID: mutation.ID, MutationNo: mutation.MutationNo, Type: mutation.Type,
			MutationDate: mutation.MutationDate.Format(dateLayout), ItemID: mutation.ItemID,
			ItemName: item.Name, Quantity: mutation.Quantity,
		}, nil
	}
	return list[0], nil
}
