package providers

import (
	"context"
	"time"
)

// MarketplaceProvider defines the interface for marketplace integrations
type MarketplaceProvider interface {
	// Platform identification
	GetPlatform() string

	// OAuth Flow
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

	// Shop Info
	GetShopInfo(ctx context.Context) (*ShopInfo, error)

	// Products
	PushProduct(ctx context.Context, product *ProductPushRequest) (*ProductPushResponse, error)
	UpdateProduct(ctx context.Context, externalID string, product *ProductUpdateRequest) error
	DeleteProduct(ctx context.Context, externalID string) error
	GetCategories(ctx context.Context) ([]ExternalCategory, error)

	// Inventory
	UpdateInventory(ctx context.Context, updates []InventoryUpdate) error
	GetInventory(ctx context.Context, externalProductIDs []string) ([]InventoryItem, error)

	// Orders
	GetOrders(ctx context.Context, params OrderQueryParams) ([]ExternalOrder, error)
	GetOrder(ctx context.Context, externalOrderID string) (*ExternalOrder, error)
	UpdateOrderStatus(ctx context.Context, externalOrderID string, status string, tracking *TrackingInfo) error

	// Webhooks
	VerifyWebhook(ctx context.Context, body []byte, headers map[string]string) (bool, error)
	ParseWebhookEvent(body []byte) (*WebhookEvent, error)
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	ShopID       string    `json:"shop_id"`
	ShopName     string    `json:"shop_name"`
}

// ShopInfo represents marketplace shop information
type ShopInfo struct {
	ShopID      string `json:"shop_id"`
	ShopName    string `json:"shop_name"`
	Status      string `json:"status"`
	Region      string `json:"region,omitempty"`
	Currency    string `json:"currency,omitempty"`
	ShopLogo    string `json:"shop_logo,omitempty"`
	Description string `json:"description,omitempty"`
}

// ProductPushRequest represents a product to push to marketplace
type ProductPushRequest struct {
	InternalID    string            `json:"internal_id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Price         float64           `json:"price"`
	OriginalPrice float64           `json:"original_price,omitempty"`
	Stock         int               `json:"stock"`
	SKU           string            `json:"sku"`
	CategoryID    string            `json:"category_id"`
	Images        []string          `json:"images"`
	Weight        float64           `json:"weight"` // in grams
	Dimensions    *Dimensions       `json:"dimensions,omitempty"`
	Variants      []VariantRequest  `json:"variants,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	Brand         string            `json:"brand,omitempty"`
	Condition     string            `json:"condition,omitempty"` // new, used
}

// Dimensions represents product dimensions
type Dimensions struct {
	Length float64 `json:"length"` // cm
	Width  float64 `json:"width"`  // cm
	Height float64 `json:"height"` // cm
}

// VariantRequest represents a product variant
type VariantRequest struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
	ImageURL string  `json:"image_url,omitempty"`
}

// ProductPushResponse represents the response from pushing a product
type ProductPushResponse struct {
	ExternalProductID string           `json:"external_product_id"`
	ExternalSKU       string           `json:"external_sku,omitempty"`
	Status            string           `json:"status"`
	VariantMappings   []VariantMapping `json:"variant_mappings,omitempty"`
	Warnings          []string         `json:"warnings,omitempty"`
}

// VariantMapping represents a mapping between internal and external variant IDs
type VariantMapping struct {
	InternalSKU string `json:"internal_sku"`
	ExternalSKU string `json:"external_sku"`
}

// ProductUpdateRequest represents a product update request
type ProductUpdateRequest struct {
	Name          string            `json:"name,omitempty"`
	Description   string            `json:"description,omitempty"`
	Price         *float64          `json:"price,omitempty"`
	OriginalPrice *float64          `json:"original_price,omitempty"`
	Stock         *int              `json:"stock,omitempty"`
	Images        []string          `json:"images,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

// ExternalCategory represents a marketplace category
type ExternalCategory struct {
	CategoryID   string             `json:"category_id"`
	CategoryName string             `json:"category_name"`
	ParentID     string             `json:"parent_id,omitempty"`
	IsLeaf       bool               `json:"is_leaf"`
	Children     []ExternalCategory `json:"children,omitempty"`
}

// InventoryUpdate represents a stock update
type InventoryUpdate struct {
	ExternalProductID string `json:"external_product_id"`
	ExternalSKU       string `json:"external_sku,omitempty"`
	Quantity          int    `json:"quantity"`
}

// InventoryItem represents inventory status from marketplace
type InventoryItem struct {
	ExternalProductID string `json:"external_product_id"`
	ExternalSKU       string `json:"external_sku"`
	Quantity          int    `json:"quantity"`
	Reserved          int    `json:"reserved,omitempty"`
}

// OrderQueryParams represents parameters for querying orders
type OrderQueryParams struct {
	Status    string     `json:"status,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Page      int        `json:"page,omitempty"`
	PageSize  int        `json:"page_size,omitempty"`
}

// ExternalOrder represents an order from marketplace
type ExternalOrder struct {
	ExternalOrderID string              `json:"external_order_id"`
	Status          string              `json:"status"`
	Items           []ExternalOrderItem `json:"items"`
	BuyerName       string              `json:"buyer_name"`
	BuyerID         string              `json:"buyer_id,omitempty"`
	ShippingAddress ShippingAddress     `json:"shipping_address"`
	TotalAmount     float64             `json:"total_amount"`
	Currency        string              `json:"currency"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	PaidAt          *time.Time          `json:"paid_at,omitempty"`
	TrackingNumber  string              `json:"tracking_number,omitempty"`
	Carrier         string              `json:"carrier,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ExternalProductID string  `json:"external_product_id"`
	SKU               string  `json:"sku"`
	Name              string  `json:"name"`
	Quantity          int     `json:"quantity"`
	Price             float64 `json:"price"`
	TotalPrice        float64 `json:"total_price"`
	VariantName       string  `json:"variant_name,omitempty"`
	ImageURL          string  `json:"image_url,omitempty"`
}

// BuyerInfo represents buyer information
type BuyerInfo struct {
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email,omitempty"`
}

// ShippingInfo represents shipping information
type ShippingInfo struct {
	RecipientName  string `json:"recipient_name"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	AddressLine2   string `json:"address_line2,omitempty"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postal_code"`
	Country        string `json:"country"`
	ShippingMethod string `json:"shipping_method,omitempty"`
}

// TrackingInfo for order fulfillment
type TrackingInfo struct {
	Courier        string     `json:"courier"`
	TrackingNumber string     `json:"tracking_number"`
	ShippedAt      *time.Time `json:"shipped_at,omitempty"`
}

// WebhookEvent represents a parsed webhook event
type WebhookEvent struct {
	Type      string      `json:"type"` // order.created, order.status_changed, etc.
	ShopID    string      `json:"shop_id"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// ProviderError represents an error from a marketplace provider
type ProviderError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Retryable  bool   `json:"retryable"`
}

func (e *ProviderError) Error() string {
	return e.Message
}

// NewProviderError creates a new ProviderError
func NewProviderError(code, message string, statusCode int, retryable bool) *ProviderError {
	return &ProviderError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Retryable:  retryable,
	}
}
