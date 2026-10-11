package inventory

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type MasterService struct {
	repo *Repository
}

func NewMasterService(repo *Repository) *MasterService {
	return &MasterService{repo: repo}
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// ---------- Kategori ----------

func toCategoryResponse(row *categoryRow) *CategoryResponse {
	return &CategoryResponse{
		ID: row.ID, Code: row.Code, Name: row.Name,
		ParentID: row.ParentID, ParentName: row.ParentName,
		Description: row.Description, CreatedAt: row.CreatedAt,
	}
}

func (s *MasterService) ListCategories(ctx context.Context, q string, page, per int64) ([]*CategoryResponse, int64, error) {
	rows, total, err := s.repo.ListCategories(ctx, q, per, (page-1)*per)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]*CategoryResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toCategoryResponse(row))
	}
	return resp, total, nil
}

func (s *MasterService) validateCategory(ctx context.Context, req SaveCategoryRequest, excludeID int64) (*Category, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("nama kategori wajib diisi")
	}
	if req.ParentID != nil && *req.ParentID > 0 {
		if *req.ParentID == excludeID && excludeID > 0 {
			return nil, errors.New("kategori tidak bisa menjadi induk dirinya sendiri")
		}
		if _, err := s.repo.FindCategory(ctx, *req.ParentID); err != nil {
			return nil, errors.New("kategori induk tidak ditemukan")
		}
	} else if req.ParentID != nil {
		req.ParentID = nil
	}
	return &Category{
		Code: cleanStr(req.Code), Name: name,
		ParentID: req.ParentID, Description: cleanStr(req.Description),
	}, nil
}

func (s *MasterService) CreateCategory(ctx context.Context, req SaveCategoryRequest) (*CategoryResponse, error) {
	category, err := s.validateCategory(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}
	return &CategoryResponse{
		ID: category.ID, Code: category.Code, Name: category.Name,
		ParentID: category.ParentID, Description: category.Description,
		CreatedAt: category.CreatedAt,
	}, nil
}

func (s *MasterService) UpdateCategory(ctx context.Context, id int64, req SaveCategoryRequest) error {
	if _, err := s.repo.FindCategory(ctx, id); err != nil {
		return mapNotFound(err)
	}
	category, err := s.validateCategory(ctx, req, id)
	if err != nil {
		return err
	}
	category.ID = id
	return s.repo.UpdateCategory(ctx, category)
}

func (s *MasterService) DeleteCategory(ctx context.Context, id int64) error {
	if _, err := s.repo.FindCategory(ctx, id); err != nil {
		return mapNotFound(err)
	}
	if count, err := s.repo.CountChildCategories(ctx, id); err != nil {
		return err
	} else if count > 0 {
		return errors.New("kategori masih memiliki sub-kategori")
	}
	if count, err := s.repo.CountItemsInCategory(ctx, id); err != nil {
		return err
	} else if count > 0 {
		return errors.New("kategori masih dipakai oleh barang")
	}
	return s.repo.SoftDeleteCategory(ctx, id)
}

// ---------- Lokasi ----------

func toLocationResponse(l *Location) *LocationResponse {
	return &LocationResponse{
		ID: l.ID, Code: l.Code, Name: l.Name,
		Building: l.Building, Floor: l.Floor,
		Description: l.Description, CreatedAt: l.CreatedAt,
	}
}

func (s *MasterService) ListLocations(ctx context.Context, q string, page, per int64) ([]*LocationResponse, int64, error) {
	rows, total, err := s.repo.ListLocations(ctx, q, per, (page-1)*per)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]*LocationResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toLocationResponse(row))
	}
	return resp, total, nil
}

func (s *MasterService) validateLocation(req SaveLocationRequest) *Location {
	name := strings.TrimSpace(req.Name)
	return &Location{
		Code: cleanStr(req.Code), Name: name,
		Building: cleanStr(req.Building), Floor: cleanStr(req.Floor),
		Description: cleanStr(req.Description),
	}
}

func (s *MasterService) CreateLocation(ctx context.Context, req SaveLocationRequest) (*LocationResponse, error) {
	location := s.validateLocation(req)
	if location.Name == "" {
		return nil, errors.New("nama lokasi wajib diisi")
	}
	if err := s.repo.CreateLocation(ctx, location); err != nil {
		return nil, err
	}
	return toLocationResponse(location), nil
}

func (s *MasterService) UpdateLocation(ctx context.Context, id int64, req SaveLocationRequest) (*LocationResponse, error) {
	if _, err := s.repo.FindLocation(ctx, id); err != nil {
		return nil, mapNotFound(err)
	}
	location := s.validateLocation(req)
	if location.Name == "" {
		return nil, errors.New("nama lokasi wajib diisi")
	}
	location.ID = id
	if err := s.repo.UpdateLocation(ctx, location); err != nil {
		return nil, err
	}
	updated, err := s.repo.FindLocation(ctx, id)
	if err != nil {
		return nil, err
	}
	return toLocationResponse(updated), nil
}

func (s *MasterService) DeleteLocation(ctx context.Context, id int64) error {
	if _, err := s.repo.FindLocation(ctx, id); err != nil {
		return mapNotFound(err)
	}
	if hasStock, err := s.repo.LocationHasStock(ctx, id); err != nil {
		return err
	} else if hasStock {
		return errors.New("lokasi masih menyimpan stok barang")
	}
	if referenced, err := s.repo.LocationReferenced(ctx, id); err != nil {
		return err
	} else if referenced {
		return errors.New("lokasi memiliki riwayat mutasi/pengadaan")
	}
	return s.repo.SoftDeleteLocation(ctx, id)
}

// ---------- Item ----------

func toItemResponse(row *itemRow) *ItemResponse {
	return &ItemResponse{
		ID: row.ID, Code: row.Code, Name: row.Name,
		CategoryID: row.CategoryID, CategoryName: row.CategoryName,
		Unit: row.Unit, MinStock: row.MinStock,
		TotalQty: row.TotalQty, BelowMin: row.TotalQty < row.MinStock && row.MinStock > 0,
		Description: row.Description, IsActive: row.IsActive, CreatedAt: row.CreatedAt,
	}
}

func (s *MasterService) ListItems(ctx context.Context, q string, categoryID *int64, active *bool, page, per int64) ([]*ItemResponse, int64, error) {
	rows, total, err := s.repo.ListItems(ctx, q, categoryID, active, per, (page-1)*per)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]*ItemResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toItemResponse(row))
	}
	return resp, total, nil
}

func (s *MasterService) validateItem(ctx context.Context, req SaveItemRequest, excludeID int64) (*Item, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" {
		return nil, errors.New("kode barang wajib diisi")
	}
	if name == "" {
		return nil, errors.New("nama barang wajib diisi")
	}
	if req.MinStock < 0 {
		return nil, errors.New("stok minimum tidak boleh negatif")
	}
	if exists, err := s.repo.ItemCodeExists(ctx, code, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("kode barang sudah digunakan")
	}

	var categoryID *int64
	if req.CategoryID != nil && *req.CategoryID > 0 {
		if _, err := s.repo.FindCategory(ctx, *req.CategoryID); err != nil {
			return nil, errors.New("kategori tidak ditemukan")
		}
		categoryID = req.CategoryID
	}

	unit := strings.TrimSpace(req.Unit)
	if unit == "" {
		unit = "buah"
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return &Item{
		Code: code, Name: name, CategoryID: categoryID,
		Unit: unit, MinStock: req.MinStock,
		Description: cleanStr(req.Description), IsActive: isActive,
	}, nil
}

func (s *MasterService) CreateItem(ctx context.Context, req SaveItemRequest) (*ItemResponse, error) {
	item, err := s.validateItem(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}
	return s.GetItem(ctx, item.ID)
}

func (s *MasterService) GetItem(ctx context.Context, id int64) (*ItemResponse, error) {
	item, err := s.repo.FindItem(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	totalQty, err := s.repo.ItemTotalStock(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := &ItemResponse{
		ID: item.ID, Code: item.Code, Name: item.Name,
		CategoryID: item.CategoryID, Unit: item.Unit, MinStock: item.MinStock,
		TotalQty: totalQty, BelowMin: totalQty < item.MinStock && item.MinStock > 0,
		Description: item.Description, IsActive: item.IsActive, CreatedAt: item.CreatedAt,
	}
	if item.CategoryID != nil {
		if cat, err := s.repo.FindCategory(ctx, *item.CategoryID); err == nil {
			resp.CategoryName = &cat.Name
		}
	}
	return resp, nil
}

func (s *MasterService) UpdateItem(ctx context.Context, id int64, req SaveItemRequest) (*ItemResponse, error) {
	if _, err := s.repo.FindItem(ctx, id); err != nil {
		return nil, mapNotFound(err)
	}
	item, err := s.validateItem(ctx, req, id)
	if err != nil {
		return nil, err
	}
	item.ID = id
	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}
	return s.GetItem(ctx, id)
}

func (s *MasterService) DeleteItem(ctx context.Context, id int64) error {
	if _, err := s.repo.FindItem(ctx, id); err != nil {
		return mapNotFound(err)
	}
	total, err := s.repo.ItemTotalStock(ctx, id)
	if err != nil {
		return err
	}
	if total > 0 {
		return errors.New("barang masih memiliki stok — lakukan mutasi keluar/penghapusan terlebih dahulu")
	}
	return s.repo.SoftDeleteItem(ctx, id)
}
