package providers

import (
	"time"
)

// OrderListParams represents parameters for listing orders
type OrderListParams struct {
	TimeFrom time.Time `json:"time_from"`
	TimeTo   time.Time `json:"time_to"`
	Status   string    `json:"status,omitempty"`
	PageSize int       `json:"page_size"`
	Cursor   string    `json:"cursor,omitempty"`
}

// ExternalOrderItem represents an order line item
type ExternalOrderItem struct {
	ExternalProductID string  `json:"external_product_id"`
	ExternalSKU       string  `json:"external_sku"`
	Name              string  `json:"name"`
	Quantity          int     `json:"quantity"`
	UnitPrice         float64 `json:"unit_price"`
	TotalPrice        float64 `json:"total_price"`
}

// ShippingAddress represents shipping address information
type ShippingAddress struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
	ZipCode string `json:"zip_code"`
}
