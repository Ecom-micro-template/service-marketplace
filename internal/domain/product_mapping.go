package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProductMapping represents the mapping between internal and external product IDs
type ProductMapping struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConnectionID      uuid.UUID  `gorm:"type:uuid;not null" json:"connection_id"`
	InternalProductID uuid.UUID  `gorm:"type:uuid;not null" json:"internal_product_id"`
	ExternalProductID string     `gorm:"type:varchar(100);not null" json:"external_product_id"`
	ExternalSKU       string     `gorm:"type:varchar(100)" json:"external_sku"`
	SyncStatus        string     `gorm:"type:varchar(50);default:'synced'" json:"sync_status"` // synced, pending, error
	LastSyncedAt      *time.Time `gorm:"type:timestamptz" json:"last_synced_at"`
	SyncError         string     `gorm:"type:text" json:"sync_error,omitempty"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Connection      *Connection      `gorm:"foreignKey:ConnectionID" json:"connection,omitempty"`
	VariantMappings []VariantMapping `gorm:"foreignKey:ProductMappingID" json:"variant_mappings,omitempty"`
}

// TableName specifies the table name for ProductMapping
func (ProductMapping) TableName() string {
	return "marketplace.product_mappings"
}

// SyncStatus constants
const (
	SyncStatusSynced  = "synced"
	SyncStatusPending = "pending"
	SyncStatusError   = "error"
)

// CreateProductMappingRequest represents a request to create a product mapping
type CreateProductMappingRequest struct {
	InternalProductID uuid.UUID `json:"internal_product_id" binding:"required"`
	ExternalProductID string    `json:"external_product_id" binding:"required"`
	ExternalSKU       string    `json:"external_sku"`
}

// UpdateProductMappingRequest represents a request to update a product mapping
type UpdateProductMappingRequest struct {
	ExternalProductID string `json:"external_product_id"`
	ExternalSKU       string `json:"external_sku"`
	SyncStatus        string `json:"sync_status"`
	SyncError         string `json:"sync_error"`
}

// ProductMappingFilter represents filter options for product mappings
type ProductMappingFilter struct {
	ConnectionID      *uuid.UUID `json:"connection_id"`
	InternalProductID *uuid.UUID `json:"internal_product_id"`
	SyncStatus        string     `json:"sync_status"`
	Page              int        `json:"page"`
	PageSize          int        `json:"page_size"`
}

// VariantMapping represents the mapping between internal and external variant IDs
type VariantMapping struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductMappingID  uuid.UUID `gorm:"type:uuid;not null" json:"product_mapping_id"`
	InternalVariantID uuid.UUID `gorm:"type:uuid;not null" json:"internal_variant_id"`
	ExternalVariantID string    `gorm:"type:varchar(100);not null" json:"external_variant_id"`
	ExternalSKU       string    `gorm:"type:varchar(100)" json:"external_sku"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	ProductMapping *ProductMapping `gorm:"foreignKey:ProductMappingID" json:"product_mapping,omitempty"`
}

// TableName specifies the table name for VariantMapping
func (VariantMapping) TableName() string {
	return "marketplace.variant_mappings"
}
