package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// OrderClient handles communication with service-order
type OrderClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewOrderClient creates a new OrderClient
func NewOrderClient(baseURL string, logger *zap.Logger) *OrderClient {
	return &OrderClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	ExternalOrderID string             `json:"external_order_id"`
	Source          string             `json:"source"` // shopee, tiktok
	CustomerName    string             `json:"customer_name"`
	CustomerEmail   string             `json:"customer_email,omitempty"`
	CustomerPhone   string             `json:"customer_phone,omitempty"`
	ShippingAddress AddressRequest     `json:"shipping_address"`
	Items           []OrderItemRequest `json:"items"`
	TotalAmount     float64            `json:"total_amount"`
	Currency        string             `json:"currency"`
	Status          string             `json:"status"`
	PaidAt          *time.Time         `json:"paid_at,omitempty"`
}

// AddressRequest represents a shipping address
type AddressRequest struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Address1   string `json:"address_1"`
	Address2   string `json:"address_2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
}

// OrderItemRequest represents an order line item
type OrderItemRequest struct {
	ProductID  string  `json:"product_id,omitempty"`
	VariantID  string  `json:"variant_id,omitempty"`
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

// CreateOrderResponse represents the response from order creation
type CreateOrderResponse struct {
	Order struct {
		ID string `json:"id"`
	} `json:"order"`
}

// CreateOrder creates an order in service-order
func (c *OrderClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (string, error) {
	url := fmt.Sprintf("%s/api/v1/orders/marketplace", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("order creation failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result CreateOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Order.ID, nil
}

// UpdateOrderStatus updates an order's status
func (c *OrderClient) UpdateOrderStatus(ctx context.Context, orderID string, status string, trackingNumber string) error {
	url := fmt.Sprintf("%s/api/v1/orders/%s/status", c.baseURL, orderID)

	body, _ := json.Marshal(map[string]string{
		"status":          status,
		"tracking_number": trackingNumber,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status update failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
