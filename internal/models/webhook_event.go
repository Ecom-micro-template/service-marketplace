package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WebhookEvent represents a webhook event received from a marketplace
type WebhookEvent struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Platform     string         `gorm:"type:varchar(50);not null" json:"platform"`
	EventType    string         `gorm:"type:varchar(100);not null" json:"event_type"`
	Payload      datatypes.JSON `gorm:"type:jsonb;not null" json:"payload"`
	Signature    string         `gorm:"type:varchar(255)" json:"signature"`
	Processed    bool           `gorm:"default:false" json:"processed"`
	ErrorMessage string         `gorm:"type:text" json:"error_message,omitempty"`
	ReceivedAt   time.Time      `gorm:"autoCreateTime" json:"received_at"`
}

// TableName specifies the table name for WebhookEvent
func (WebhookEvent) TableName() string {
	return "marketplace.webhook_events"
}

// Event type constants for Shopee
const (
	// Shopee event types
	ShopeeEventOrderCreated         = "shopee.order.created"
	ShopeeEventOrderStatusChanged   = "shopee.order.status_changed"
	ShopeeEventOrderShipped         = "shopee.order.shipped"
	ShopeeEventOrderCompleted       = "shopee.order.completed"
	ShopeeEventOrderCancelled       = "shopee.order.cancelled"
	ShopeeEventProductBanned        = "shopee.product.banned"
	ShopeeEventProductUnbanned      = "shopee.product.unbanned"
	ShopeeEventInventoryChanged     = "shopee.inventory.changed"
	ShopeeEventAuthorizationRevoked = "shopee.authorization.revoked"
)

// Event type constants for TikTok
const (
	// TikTok event types
	TikTokEventOrderCreated       = "tiktok.order.created"
	TikTokEventOrderStatusChanged = "tiktok.order.status_changed"
	TikTokEventOrderShipped       = "tiktok.order.shipped"
	TikTokEventOrderCompleted     = "tiktok.order.completed"
	TikTokEventOrderCancelled     = "tiktok.order.cancelled"
	TikTokEventProductCreated     = "tiktok.product.created"
	TikTokEventProductUpdated     = "tiktok.product.updated"
	TikTokEventProductDeleted     = "tiktok.product.deleted"
	TikTokEventInventoryUpdated   = "tiktok.inventory.updated"
)

// WebhookEventFilter represents filter options for webhook events
type WebhookEventFilter struct {
	Platform  string     `json:"platform"`
	EventType string     `json:"event_type"`
	Processed *bool      `json:"processed"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}

// MarkProcessedRequest represents a request to mark events as processed
type MarkProcessedRequest struct {
	EventIDs     []uuid.UUID `json:"event_ids" binding:"required"`
	ErrorMessage string      `json:"error_message,omitempty"`
}
