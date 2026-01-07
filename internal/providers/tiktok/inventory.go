package tiktok

import (
	"context"
	"fmt"
	"net/http"

	"github.com/niaga-platform/service-marketplace/internal/providers"
)

// InventoryProvider implements inventory operations for TikTok Shop
type InventoryProvider struct {
	client *Client
}

// NewInventoryProvider creates a new TikTok inventory provider
func NewInventoryProvider(client *Client) *InventoryProvider {
	return &InventoryProvider{client: client}
}

// UpdateStock updates stock for a single product
func (p *InventoryProvider) UpdateStock(ctx context.Context, productID, skuID string, quantity int) error {
	req := &Request{
		Method: http.MethodPut,
		Path:   UpdateInventoryPath,
		Body: map[string]interface{}{
			"skus": []map[string]interface{}{
				{
					"product_id": productID,
					"id":         skuID,
					"stock_infos": []map[string]interface{}{
						{"available_stock": quantity},
					},
				},
			},
		},
		NeedAuth: true,
	}

	var resp BaseResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	if resp.HasError() {
		return fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	return nil
}

// UpdateBatchStock updates stock for multiple products
func (p *InventoryProvider) UpdateBatchStock(ctx context.Context, updates []providers.InventoryUpdate) ([]providers.InventoryUpdateResult, error) {
	results := make([]providers.InventoryUpdateResult, len(updates))

	for i, update := range updates {
		err := p.UpdateStock(ctx, update.ExternalProductID, update.ExternalSKU, update.Quantity)
		results[i] = providers.InventoryUpdateResult{
			ExternalProductID: update.ExternalProductID,
			Success:           err == nil,
		}
		if err != nil {
			results[i].Error = err.Error()
		}
	}

	return results, nil
}

// GetStock fetches current stock levels
func (p *InventoryProvider) GetStock(ctx context.Context, productIDs []string) ([]providers.InventoryItem, error) {
	req := &Request{
		Method: http.MethodPost,
		Path:   GetProductsPath,
		Body: map[string]interface{}{
			"product_ids": productIDs,
		},
		NeedAuth: true,
	}

	var resp struct {
		BaseResponse
		Data struct {
			Products []struct {
				ProductID string `json:"product_id"`
				SKUs      []struct {
					ID         string `json:"id"`
					SellerSKU  string `json:"seller_sku"`
					StockInfos []struct {
						AvailableStock int `json:"available_stock"`
					} `json:"stock_infos"`
				} `json:"skus"`
			} `json:"products"`
		} `json:"data"`
	}

	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	items := make([]providers.InventoryItem, 0)
	for _, product := range resp.Data.Products {
		for _, sku := range product.SKUs {
			stock := 0
			if len(sku.StockInfos) > 0 {
				stock = sku.StockInfos[0].AvailableStock
			}
			items = append(items, providers.InventoryItem{
				ExternalProductID: product.ProductID,
				ExternalSKU:       sku.ID,
				Quantity:          stock,
			})
		}
	}

	return items, nil
}
