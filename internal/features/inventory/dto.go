package inventory

import (
	"time"

	"penggajian/server/internal/platform/httpx"
)

type CategoryResponse struct {
	ID          int64     `json:"id"`
	Code        *string   `json:"code"`
	Name        string    `json:"name"`
	ParentID    *int64    `json:"parent_id"`
	ParentName  *string   `json:"parent_name,omitempty"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type SaveCategoryRequest struct {
	Code        *string `json:"code"`
	Name        string  `json:"name"`
	ParentID    *int64  `json:"parent_id"`
	Description *string `json:"description"`
}

type LocationResponse struct {
	ID          int64     `json:"id"`
	Code        *string   `json:"code"`
	Name        string    `json:"name"`
	Building    *string   `json:"building"`
	Floor       *string   `json:"floor"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type SaveLocationRequest struct {
	Code        *string `json:"code"`
	Name        string  `json:"name"`
	Building    *string `json:"building"`
	Floor       *string `json:"floor"`
	Description *string `json:"description"`
}

type ItemResponse struct {
	ID           int64     `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	CategoryID   *int64    `json:"category_id"`
	CategoryName *string   `json:"category_name,omitempty"`
	Unit         string    `json:"unit"`
	MinStock     int       `json:"min_stock"`
	TotalQty     int       `json:"total_qty"`
	BelowMin     bool      `json:"below_min"`
	Description  *string   `json:"description"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type SaveItemRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	CategoryID  *int64  `json:"category_id"`
	Unit        string  `json:"unit"`
	MinStock    int     `json:"min_stock"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type StockResponse struct {
	ID           int64  `json:"id"`
	ItemID       int64  `json:"item_id"`
	ItemCode     string `json:"item_code"`
	ItemName     string `json:"item_name"`
	Unit         string `json:"unit"`
	LocationID   int64  `json:"location_id"`
	LocationName string `json:"location_name"`
	Condition    string `json:"condition"`
	Quantity     int    `json:"quantity"`
}

type StockSummaryRow struct {
	ItemResponse
	QtyBaik   int `json:"qty_baik" gorm:"column:qty_baik"`
	QtyRingan int `json:"qty_rusak_ringan" gorm:"column:qty_ringan"`
	QtyBerat  int `json:"qty_rusak_berat" gorm:"column:qty_berat"`
}

type MutationResponse struct {
	ID             int64      `json:"id"`
	MutationNo     string     `json:"mutation_no"`
	Type           string     `json:"type"`
	MutationDate   string     `json:"mutation_date"`
	ItemID         int64      `json:"item_id"`
	ItemCode       string     `json:"item_code"`
	ItemName       string     `json:"item_name"`
	FromLocationID *int64     `json:"from_location_id"`
	FromLocation   *string    `json:"from_location"`
	ToLocationID   *int64     `json:"to_location_id"`
	ToLocation     *string    `json:"to_location"`
	FromCondition  *string    `json:"from_condition"`
	ToCondition    *string    `json:"to_condition"`
	Quantity       int        `json:"quantity"`
	ReferenceNo    *string    `json:"reference_no"`
	Note           *string    `json:"note"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateMutationRequest struct {
	Type           string  `json:"type"`
	MutationDate   string  `json:"mutation_date"`
	ItemID         int64   `json:"item_id"`
	FromLocationID *int64  `json:"from_location_id"`
	ToLocationID   *int64  `json:"to_location_id"`
	FromCondition  *string `json:"from_condition"`
	ToCondition    *string `json:"to_condition"`
	Quantity       int     `json:"quantity"`
	ReferenceNo    *string `json:"reference_no"`
	Note           *string `json:"note"`
}

type PurchaseResponse struct {
	ID           int64      `json:"id"`
	PurchaseNo   string     `json:"purchase_no"`
	PurchaseDate string     `json:"purchase_date"`
	SupplierName *string    `json:"supplier_name"`
	Description  *string    `json:"description"`
	TotalCost    int64      `json:"total_cost"`
	Status       string     `json:"status"`
	ItemCount    int        `json:"item_count"`
	CreatedAt    time.Time  `json:"created_at"`
}

type PurchaseItemResponse struct {
	ID           int64  `json:"id"`
	ItemID       int64  `json:"item_id"`
	ItemCode     string `json:"item_code"`
	ItemName     string `json:"item_name"`
	Unit         string `json:"unit"`
	Quantity     int    `json:"quantity"`
	UnitPrice    int64  `json:"unit_price"`
	Subtotal     int64  `json:"subtotal"`
	LocationID   int64  `json:"location_id"`
	LocationName string `json:"location_name"`
}

type PurchaseDetailResponse struct {
	PurchaseResponse
	Items []*PurchaseItemResponse `json:"items"`
}

type PurchaseItemInput struct {
	ItemID     int64 `json:"item_id"`
	Quantity   int   `json:"quantity"`
	UnitPrice  int64 `json:"unit_price"`
	LocationID int64 `json:"location_id"`
}

type SavePurchaseRequest struct {
	PurchaseDate string              `json:"purchase_date"`
	SupplierName string              `json:"supplier_name"`
	Description  string              `json:"description"`
	Items        []PurchaseItemInput `json:"items"`
}

type ListResult[T any] struct {
	Data []*T        `json:"data"`
	Meta *httpx.Meta `json:"meta"`
}
