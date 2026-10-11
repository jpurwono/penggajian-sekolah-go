package inventory

import (
	"time"

	"gorm.io/gorm"
)

const (
	MutIn       = "in"
	MutOut      = "out"
	MutTransfer = "transfer"
	MutRepair   = "repair"
	MutDiscard  = "discard"

	CondBaik       = "baik"
	CondRusakRingan = "rusak_ringan"
	CondRusakBerat  = "rusak_berat"
)

func IsValidMutationType(t string) bool {
	return t == MutIn || t == MutOut || t == MutTransfer || t == MutRepair || t == MutDiscard
}

func IsValidCondition(c string) bool {
	return c == CondBaik || c == CondRusakRingan || c == CondRusakBerat
}

type Category struct {
	ID          int64          `gorm:"primaryKey"`
	Code        *string        `gorm:"column:code"`
	Name        string         `gorm:"column:name"`
	ParentID    *int64         `gorm:"column:parent_id"`
	Description *string        `gorm:"column:description"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Category) TableName() string { return "inventory_categories" }

type Location struct {
	ID          int64          `gorm:"primaryKey"`
	Code        *string        `gorm:"column:code"`
	Name        string         `gorm:"column:name"`
	Building    *string        `gorm:"column:building"`
	Floor       *string        `gorm:"column:floor"`
	Description *string        `gorm:"column:description"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Location) TableName() string { return "inventory_locations" }

type Item struct {
	ID          int64          `gorm:"primaryKey"`
	Code        string         `gorm:"column:code"`
	Name        string         `gorm:"column:name"`
	CategoryID  *int64         `gorm:"column:category_id"`
	Unit        string         `gorm:"column:unit"`
	MinStock    int            `gorm:"column:min_stock"`
	Description *string        `gorm:"column:description"`
	IsActive    bool           `gorm:"column:is_active"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Item) TableName() string { return "inventory_items" }

type Stock struct {
	ID         int64     `gorm:"primaryKey"`
	ItemID     int64     `gorm:"column:item_id"`
	LocationID int64     `gorm:"column:location_id"`
	Condition  string    `gorm:"column:condition"`
	Quantity   int       `gorm:"column:quantity"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (Stock) TableName() string { return "inventory_stocks" }

type Mutation struct {
	ID              int64     `gorm:"primaryKey"`
	MutationNo      string    `gorm:"column:mutation_no"`
	Type            string    `gorm:"column:type"`
	MutationDate    time.Time `gorm:"column:mutation_date"`
	ItemID          int64     `gorm:"column:item_id"`
	FromLocationID  *int64    `gorm:"column:from_location_id"`
	ToLocationID    *int64    `gorm:"column:to_location_id"`
	FromCondition   *string   `gorm:"column:from_condition"`
	ToCondition     *string   `gorm:"column:to_condition"`
	Quantity        int       `gorm:"column:quantity"`
	ReferenceNo     *string   `gorm:"column:reference_no"`
	Note            *string   `gorm:"column:note"`
	CreatedBy       *int64    `gorm:"column:created_by"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (Mutation) TableName() string { return "inventory_mutations" }

const (
	PurchaseDraft    = "draft"
	PurchaseReceived = "received"
)

type Purchase struct {
	ID           int64     `gorm:"primaryKey"`
	PurchaseNo   string    `gorm:"column:purchase_no"`
	PurchaseDate time.Time `gorm:"column:purchase_date"`
	SupplierName *string   `gorm:"column:supplier_name"`
	Description  *string   `gorm:"column:description"`
	TotalCost    int64     `gorm:"column:total_cost"`
	Status       string    `gorm:"column:status"`
	CreatedBy    *int64    `gorm:"column:created_by"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Purchase) TableName() string { return "inventory_purchases" }

type PurchaseItem struct {
	ID         int64     `gorm:"primaryKey"`
	PurchaseID int64     `gorm:"column:purchase_id"`
	ItemID     int64     `gorm:"column:item_id"`
	Quantity   int       `gorm:"column:quantity"`
	UnitPrice  int64     `gorm:"column:unit_price"`
	Subtotal   int64     `gorm:"column:subtotal"`
	LocationID int64     `gorm:"column:location_id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (PurchaseItem) TableName() string { return "inventory_purchase_items" }
