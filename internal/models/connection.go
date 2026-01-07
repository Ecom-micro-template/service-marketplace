package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Connection represents a marketplace connection (OAuth credentials)
type Connection struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Platform       string         `gorm:"type:varchar(50);not null" json:"platform"` // 'shopee' or 'tiktok'
	ShopID         string         `gorm:"type:varchar(100);not null" json:"shop_id"`
	ShopName       string         `gorm:"type:varchar(255)" json:"shop_name"`
	AccessToken    string         `gorm:"type:text;not null" json:"-"` // Encrypted, hidden from JSON
	RefreshToken   string         `gorm:"type:text" json:"-"`          // Encrypted, hidden from JSON
	TokenExpiresAt *time.Time     `gorm:"type:timestamptz" json:"token_expires_at"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	Settings       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	ProductMappings  []ProductMapping   `gorm:"foreignKey:ConnectionID" json:"product_mappings,omitempty"`
	CategoryMappings []CategoryMapping  `gorm:"foreignKey:ConnectionID" json:"category_mappings,omitempty"`
	Orders           []MarketplaceOrder `gorm:"foreignKey:ConnectionID" json:"orders,omitempty"`
	SyncJobs         []SyncJob          `gorm:"foreignKey:ConnectionID" json:"sync_jobs,omitempty"`
}

// TableName specifies the table name for Connection
func (Connection) TableName() string {
	return "marketplace.connections"
}

// ConnectionResponse represents a connection response without sensitive data
type ConnectionResponse struct {
	ID             uuid.UUID      `json:"id"`
	Platform       string         `json:"platform"`
	ShopID         string         `json:"shop_id"`
	ShopName       string         `json:"shop_name"`
	TokenExpiresAt *time.Time     `json:"token_expires_at"`
	IsActive       bool           `json:"is_active"`
	Settings       datatypes.JSON `json:"settings"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// ToResponse converts Connection to ConnectionResponse
func (c *Connection) ToResponse() *ConnectionResponse {
	return &ConnectionResponse{
		ID:             c.ID,
		Platform:       c.Platform,
		ShopID:         c.ShopID,
		ShopName:       c.ShopName,
		TokenExpiresAt: c.TokenExpiresAt,
		IsActive:       c.IsActive,
		Settings:       c.Settings,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

// CreateConnectionRequest represents a request to create a connection
type CreateConnectionRequest struct {
	Platform     string `json:"platform" binding:"required,oneof=shopee tiktok"`
	ShopID       string `json:"shop_id" binding:"required"`
	ShopName     string `json:"shop_name"`
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until token expires
}
