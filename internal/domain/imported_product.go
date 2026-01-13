package domain

import (
	"time"

	"github.com/google/uuid"
)

// ImportedProduct represents a product imported from a marketplace (not yet mapped to internal product)
type ImportedProduct struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConnectionID      uuid.UUID  `gorm:"type:uuid;not null" json:"connection_id"`
	ExternalProductID string     `gorm:"type:varchar(100);not null" json:"external_product_id"`
	ExternalSKU       string     `gorm:"type:varchar(100)" json:"external_sku"`
	Name              string     `gorm:"type:varchar(500);not null" json:"name"`
	Description       string     `gorm:"type:text" json:"description"`
	Price             float64    `gorm:"type:decimal(12,2)" json:"price"`
	Stock             int        `gorm:"default:0" json:"stock"`
	CategoryID        string     `gorm:"type:varchar(100)" json:"category_id"`
	Status            string     `gorm:"type:varchar(50)" json:"status"` // NORMAL, BANNED, etc.
	ImageURL          string     `gorm:"type:text" json:"image_url"`     // First image URL
	IsMapped          bool       `gorm:"default:false" json:"is_mapped"` // Whether this product is mapped to an internal product
	MappedToProductID *uuid.UUID `gorm:"type:uuid" json:"mapped_to_product_id,omitempty"`
	ImportedAt        time.Time  `gorm:"autoCreateTime" json:"imported_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Connection *Connection `gorm:"foreignKey:ConnectionID" json:"connection,omitempty"`
}

// TableName specifies the table name for ImportedProduct
func (ImportedProduct) TableName() string {
	return "marketplace.imported_products"
}

// ImportedProductFilter represents filter options for imported products
type ImportedProductFilter struct {
	ConnectionID *uuid.UUID `json:"connection_id"`
	IsMapped     *bool      `json:"is_mapped"`
	Status       string     `json:"status"`
	Search       string     `json:"search"`
	Page         int        `json:"page"`
	PageSize     int        `json:"page_size"`
}
