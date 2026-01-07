package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// MarketplaceOrder represents an order received from a marketplace
type MarketplaceOrder struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConnectionID    uuid.UUID      `gorm:"type:uuid" json:"connection_id"`
	InternalOrderID *uuid.UUID     `gorm:"type:uuid" json:"internal_order_id"` // Linked order in service-order
	ExternalOrderID string         `gorm:"type:varchar(100);not null" json:"external_order_id"`
	Platform        string         `gorm:"type:varchar(50);not null" json:"platform"`
	Status          string         `gorm:"type:varchar(50);not null" json:"status"` // Platform-specific status
	OrderData       datatypes.JSON `gorm:"type:jsonb;not null" json:"order_data"`
	ShippingInfo    datatypes.JSON `gorm:"type:jsonb" json:"shipping_info"`
	BuyerInfo       datatypes.JSON `gorm:"type:jsonb" json:"buyer_info"`
	TotalAmount     float64        `gorm:"type:decimal(12,2)" json:"total_amount"`
	Currency        string         `gorm:"type:varchar(10);default:'MYR'" json:"currency"`
	SyncedAt        *time.Time     `gorm:"type:timestamptz" json:"synced_at"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Connection *Connection `gorm:"foreignKey:ConnectionID" json:"connection,omitempty"`
}

// TableName specifies the table name for MarketplaceOrder
func (MarketplaceOrder) TableName() string {
	return "marketplace.orders"
}

// Order status constants (platform-specific statuses will vary)
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
	OrderStatusRefunded  = "refunded"
	OrderStatusReturned  = "returned"
)

// OrderDataJSON represents the structure stored in order_data
type OrderDataJSON struct {
	Items           []OrderItemJSON `json:"items"`
	SubtotalAmount  float64         `json:"subtotal_amount"`
	ShippingFee     float64         `json:"shipping_fee"`
	DiscountAmount  float64         `json:"discount_amount"`
	PlatformFee     float64         `json:"platform_fee"`
	PaymentMethod   string          `json:"payment_method"`
	Notes           string          `json:"notes"`
	MarketplaceData interface{}     `json:"marketplace_data"` // Platform-specific data
}

// OrderItemJSON represents an item in the order
type OrderItemJSON struct {
	ExternalProductID string  `json:"external_product_id"`
	InternalProductID string  `json:"internal_product_id,omitempty"`
	SKU               string  `json:"sku"`
	Name              string  `json:"name"`
	Quantity          int     `json:"quantity"`
	UnitPrice         float64 `json:"unit_price"`
	TotalPrice        float64 `json:"total_price"`
	VariantName       string  `json:"variant_name,omitempty"`
}

// BuyerInfoJSON represents buyer information
type BuyerInfoJSON struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	UserID string `json:"user_id,omitempty"`
}

// ShippingInfoJSON represents shipping information
type ShippingInfoJSON struct {
	RecipientName  string `json:"recipient_name"`
	Phone          string `json:"phone"`
	AddressLine1   string `json:"address_line1"`
	AddressLine2   string `json:"address_line2,omitempty"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postal_code"`
	Country        string `json:"country"`
	Courier        string `json:"courier,omitempty"`
	TrackingNumber string `json:"tracking_number,omitempty"`
}

// MarketplaceOrderFilter represents filter options for marketplace orders
type MarketplaceOrderFilter struct {
	ConnectionID    *uuid.UUID `json:"connection_id"`
	Platform        string     `json:"platform"`
	Status          string     `json:"status"`
	ExternalOrderID string     `json:"external_order_id"`
	ImportedOnly    *bool      `json:"imported_only"` // Only orders with internal_order_id
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	Page            int        `json:"page"`
	PageSize        int        `json:"page_size"`
}

// ImportOrderRequest represents a request to import an order into the internal system
type ImportOrderRequest struct {
	MarketplaceOrderID uuid.UUID `json:"marketplace_order_id" binding:"required"`
}

// UpdateOrderStatusRequest represents a request to update order status
type UpdateOrderStatusRequest struct {
	Status         string `json:"status" binding:"required"`
	TrackingNumber string `json:"tracking_number"`
	Courier        string `json:"courier"`
}
