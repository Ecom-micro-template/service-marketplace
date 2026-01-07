package shopee

import (
	"context"
	"fmt"
	"net/http"

	"github.com/niaga-platform/service-marketplace/internal/providers"
)

// InventoryProvider implements inventory operations for Shopee
type InventoryProvider struct {
	client *Client
}

// NewInventoryProvider creates a new Shopee inventory provider
func NewInventoryProvider(client *Client) *InventoryProvider {
	return &InventoryProvider{client: client}
}

// UpdateStock updates stock for a single product
func (p *InventoryProvider) UpdateStock(ctx context.Context, externalProductID string, quantity int) error {
	req := &Request{
		Method: http.MethodPost,
		Path:   UpdateStockPath,
		Body: map[string]interface{}{
			"item_id": externalProductID,
			"stock_list": []map[string]interface{}{
				{
					"model_id":     0,
					"normal_stock": quantity,
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
		return fmt.Errorf("shopee error: %s", resp.GetError())
	}

	return nil
}

// UpdateBatchStock updates stock for multiple products
func (p *InventoryProvider) UpdateBatchStock(ctx context.Context, updates []providers.InventoryUpdate) ([]providers.InventoryUpdateResult, error) {
	results := make([]providers.InventoryUpdateResult, len(updates))

	for i, update := range updates {
		err := p.UpdateStock(ctx, update.ExternalProductID, update.Quantity)
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
func (p *InventoryProvider) GetStock(ctx context.Context, externalProductIDs []string) ([]providers.InventoryItem, error) {
	itemIDList := ""
	for i, id := range externalProductIDs {
		if i > 0 {
			itemIDList += ","
		}
		itemIDList += id
	}

	req := &Request{
		Method: http.MethodGet,
		Path:   GetItemInfoPath,
		Query: map[string]string{
			"item_id_list": itemIDList,
		},
		NeedAuth: true,
	}

	var resp struct {
		BaseResponse
		Response struct {
			ItemList []struct {
				ItemID      int64 `json:"item_id"`
				StockInfoV2 struct {
					SummaryInfo struct {
						TotalAvailableStock int `json:"total_available_stock"`
						TotalReservedStock  int `json:"total_reserved_stock"`
					} `json:"summary_info"`
				} `json:"stock_info_v2"`
			} `json:"item_list"`
		} `json:"response"`
	}

	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("shopee error: %s", resp.GetError())
	}

	items := make([]providers.InventoryItem, len(resp.Response.ItemList))
	for i, item := range resp.Response.ItemList {
		items[i] = providers.InventoryItem{
			ExternalProductID: fmt.Sprintf("%d", item.ItemID),
			Quantity:          item.StockInfoV2.SummaryInfo.TotalAvailableStock,
			Reserved:          item.StockInfoV2.SummaryInfo.TotalReservedStock,
		}
	}

	return items, nil
}
