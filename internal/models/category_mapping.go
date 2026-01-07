package models

import (
	"time"

	"github.com/google/uuid"
)

// CategoryMapping represents the mapping between internal and external category IDs
type CategoryMapping struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConnectionID         uuid.UUID `gorm:"type:uuid;not null" json:"connection_id"`
	InternalCategoryID   uuid.UUID `gorm:"type:uuid;not null" json:"internal_category_id"`
	ExternalCategoryID   string    `gorm:"type:varchar(100);not null" json:"external_category_id"`
	ExternalCategoryName string    `gorm:"type:varchar(255)" json:"external_category_name"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Connection *Connection `gorm:"foreignKey:ConnectionID" json:"connection,omitempty"`
}

// TableName specifies the table name for CategoryMapping
func (CategoryMapping) TableName() string {
	return "marketplace.category_mappings"
}

// CreateCategoryMappingRequest represents a request to create a category mapping
type CreateCategoryMappingRequest struct {
	InternalCategoryID   uuid.UUID `json:"internal_category_id" binding:"required"`
	ExternalCategoryID   string    `json:"external_category_id" binding:"required"`
	ExternalCategoryName string    `json:"external_category_name"`
}

// CategoryMappingFilter represents filter options for category mappings
type CategoryMappingFilter struct {
	ConnectionID       *uuid.UUID `json:"connection_id"`
	InternalCategoryID *uuid.UUID `json:"internal_category_id"`
}

// ExternalCategoryResponse represents an external marketplace category
type ExternalCategoryResponse struct {
	CategoryID   string                     `json:"category_id"`
	CategoryName string                     `json:"category_name"`
	ParentID     string                     `json:"parent_id,omitempty"`
	HasChildren  bool                       `json:"has_children"`
	Children     []ExternalCategoryResponse `json:"children,omitempty"`
}
