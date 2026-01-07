package tiktok

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/niaga-platform/service-marketplace/internal/providers"
)

const (
	GetOrderListPath   = "/api/orders/search"
	GetOrderDetailPath = "/api/orders/detail/query"
	ShipOrderPath      = "/api/fulfillment/package/ship"
)

// OrderProvider implements order operations for TikTok Shop
type OrderProvider struct {
	client *Client
}

// NewOrderProvider creates a new TikTok order provider
func NewOrderProvider(client *Client) *OrderProvider {
	return &OrderProvider{client: client}
}

// GetOrders fetches orders from TikTok
func (p *OrderProvider) GetOrders(ctx context.Context, params *providers.OrderListParams) ([]providers.ExternalOrder, string, error) {
	body := map[string]interface{}{
		"create_time_ge": params.TimeFrom.Unix(),
		"create_time_lt": params.TimeTo.Unix(),
		"page_size":      params.PageSize,
	}

	if params.Cursor != "" {
		body["cursor"] = params.Cursor
	}
	if params.Status != "" {
		body["order_status"] = p.mapStatusToAPI(params.Status)
	}

	req := &Request{
		Method:   http.MethodPost,
		Path:     GetOrderListPath,
		Body:     body,
		NeedAuth: true,
	}

	var resp struct {
		BaseResponse
		Data struct {
			NextCursor string `json:"next_cursor"`
			Orders     []struct {
				OrderID          string `json:"order_id"`
				OrderStatus      int    `json:"order_status"`
				CreateTime       int64  `json:"create_time"`
				UpdateTime       int64  `json:"update_time"`
				PaymentMethod    string `json:"payment_method_name"`
				TotalAmount      string `json:"total_amount"`
				Currency         string `json:"currency"`
				RecipientAddress struct {
					Name         string `json:"name"`
					PhoneNumber  string `json:"phone_number"`
					AddressLine1 string `json:"address_line1"`
					City         string `json:"city"`
					State        string `json:"state"`
					PostalCode   string `json:"postal_code"`
					Region       string `json:"region_code"`
				} `json:"recipient_address"`
				LineItems []struct {
					SkuID       string `json:"sku_id"`
					ProductID   string `json:"product_id"`
					ProductName string `json:"product_name"`
					SkuName     string `json:"sku_name"`
					Quantity    int    `json:"quantity"`
					SalePrice   string `json:"sale_price"`
				} `json:"line_items"`
				TrackingNumber   string `json:"tracking_number"`
				ShippingProvider string `json:"shipping_provider"`
			} `json:"orders"`
		} `json:"data"`
	}

	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, "", fmt.Errorf("failed to get orders: %w", err)
	}

	if resp.HasError() {
		return nil, "", fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	orders := make([]providers.ExternalOrder, len(resp.Data.Orders))
	for i, o := range resp.Data.Orders {
		items := make([]providers.ExternalOrderItem, len(o.LineItems))
		for j, item := range o.LineItems {
			price := parseFloat(item.SalePrice)
			items[j] = providers.ExternalOrderItem{
				ExternalProductID: item.ProductID,
				ExternalSKU:       item.SkuID,
				Name:              item.ProductName,
				Quantity:          item.Quantity,
				UnitPrice:         price,
				TotalPrice:        price * float64(item.Quantity),
			}
		}

		orders[i] = providers.ExternalOrder{
			ExternalOrderID: o.OrderID,
			Status:          p.mapOrderStatus(o.OrderStatus),
			TotalAmount:     parseFloat(o.TotalAmount),
			Currency:        o.Currency,
			CreatedAt:       time.Unix(o.CreateTime, 0),
			UpdatedAt:       time.Unix(o.UpdateTime, 0),
			ShippingAddress: providers.ShippingAddress{
				Name:    o.RecipientAddress.Name,
				Phone:   o.RecipientAddress.PhoneNumber,
				City:    o.RecipientAddress.City,
				State:   o.RecipientAddress.State,
				Country: o.RecipientAddress.Region,
				ZipCode: o.RecipientAddress.PostalCode,
				Address: o.RecipientAddress.AddressLine1,
			},
			Items:          items,
			TrackingNumber: o.TrackingNumber,
			Carrier:        o.ShippingProvider,
		}
	}

	return orders, resp.Data.NextCursor, nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func (p *OrderProvider) mapOrderStatus(status int) string {
	switch status {
	case 100:
		return "pending_payment"
	case 111, 112:
		return "pending_shipment"
	case 121:
		return "shipped"
	case 130:
		return "completed"
	case 140:
		return "cancelled"
	default:
		return fmt.Sprintf("unknown_%d", status)
	}
}

func (p *OrderProvider) mapStatusToAPI(status string) int {
	switch status {
	case "pending_payment":
		return 100
	case "pending_shipment":
		return 111
	case "shipped":
		return 121
	case "completed":
		return 130
	case "cancelled":
		return 140
	default:
		return 0
	}
}

// GetOrder fetches a single order
func (p *OrderProvider) GetOrder(ctx context.Context, orderID string) (*providers.ExternalOrder, error) {
	req := &Request{
		Method: http.MethodPost,
		Path:   GetOrderDetailPath,
		Body: map[string]interface{}{
			"order_id": orderID,
		},
		NeedAuth: true,
	}

	var resp struct {
		BaseResponse
		Data struct {
			OrderID          string `json:"order_id"`
			OrderStatus      int    `json:"order_status"`
			CreateTime       int64  `json:"create_time"`
			UpdateTime       int64  `json:"update_time"`
			TotalAmount      string `json:"total_amount"`
			Currency         string `json:"currency"`
			RecipientAddress struct {
				Name         string `json:"name"`
				PhoneNumber  string `json:"phone_number"`
				AddressLine1 string `json:"address_line1"`
				City         string `json:"city"`
				State        string `json:"state"`
				PostalCode   string `json:"postal_code"`
				Region       string `json:"region_code"`
			} `json:"recipient_address"`
			LineItems []struct {
				SkuID       string `json:"sku_id"`
				ProductID   string `json:"product_id"`
				ProductName string `json:"product_name"`
				Quantity    int    `json:"quantity"`
				SalePrice   string `json:"sale_price"`
			} `json:"line_items"`
			TrackingNumber   string `json:"tracking_number"`
			ShippingProvider string `json:"shipping_provider"`
		} `json:"data"`
	}

	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	o := resp.Data
	items := make([]providers.ExternalOrderItem, len(o.LineItems))
	for j, item := range o.LineItems {
		price := parseFloat(item.SalePrice)
		items[j] = providers.ExternalOrderItem{
			ExternalProductID: item.ProductID,
			ExternalSKU:       item.SkuID,
			Name:              item.ProductName,
			Quantity:          item.Quantity,
			UnitPrice:         price,
			TotalPrice:        price * float64(item.Quantity),
		}
	}

	return &providers.ExternalOrder{
		ExternalOrderID: o.OrderID,
		Status:          p.mapOrderStatus(o.OrderStatus),
		TotalAmount:     parseFloat(o.TotalAmount),
		Currency:        o.Currency,
		CreatedAt:       time.Unix(o.CreateTime, 0),
		UpdatedAt:       time.Unix(o.UpdateTime, 0),
		ShippingAddress: providers.ShippingAddress{
			Name:    o.RecipientAddress.Name,
			Phone:   o.RecipientAddress.PhoneNumber,
			City:    o.RecipientAddress.City,
			State:   o.RecipientAddress.State,
			Country: o.RecipientAddress.Region,
			ZipCode: o.RecipientAddress.PostalCode,
			Address: o.RecipientAddress.AddressLine1,
		},
		Items:          items,
		TrackingNumber: o.TrackingNumber,
		Carrier:        o.ShippingProvider,
	}, nil
}

// UpdateOrderStatus updates order status
func (p *OrderProvider) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	if status == "shipped" {
		req := &Request{
			Method: http.MethodPost,
			Path:   ShipOrderPath,
			Body: map[string]interface{}{
				"order_id": orderID,
			},
			NeedAuth: true,
		}

		var resp BaseResponse
		if err := p.client.Do(ctx, req, &resp); err != nil {
			return fmt.Errorf("failed to ship order: %w", err)
		}

		if resp.HasError() {
			return fmt.Errorf("tiktok error: %s", resp.GetError())
		}
	}

	return nil
}
