package tiktok

import (
	"context"
	"fmt"
	"net/http"

	"github.com/niaga-platform/service-marketplace/internal/providers"
)

const (
	// Product API paths
	CreateProductPath   = "/api/products"
	UpdateProductPath   = "/api/products"
	DeleteProductPath   = "/api/products"
	GetCategoriesPath   = "/api/products/categories"
	UpdateInventoryPath = "/api/products/stocks"
	GetProductsPath     = "/api/products/search"
)

// ProductProvider implements product operations for TikTok Shop
type ProductProvider struct {
	client *Client
}

// NewProductProvider creates a new TikTok product provider
func NewProductProvider(client *Client) *ProductProvider {
	return &ProductProvider{client: client}
}

// GetCategories fetches marketplace categories
func (p *ProductProvider) GetCategories(ctx context.Context) ([]providers.ExternalCategory, error) {
	req := &Request{
		Method:   http.MethodGet,
		Path:     GetCategoriesPath,
		NeedAuth: true,
	}

	var resp struct {
		BaseResponse
		Data struct {
			Categories []struct {
				ID        string `json:"id"`
				ParentID  string `json:"parent_id"`
				LocalName string `json:"local_name"`
				IsLeaf    bool   `json:"is_leaf"`
			} `json:"categories"`
		} `json:"data"`
	}

	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	categories := make([]providers.ExternalCategory, len(resp.Data.Categories))
	for i, cat := range resp.Data.Categories {
		categories[i] = providers.ExternalCategory{
			CategoryID:   cat.ID,
			CategoryName: cat.LocalName,
			ParentID:     cat.ParentID,
			IsLeaf:       cat.IsLeaf,
		}
	}

	return categories, nil
}

// PushProduct creates a new product on TikTok Shop
func (p *ProductProvider) PushProduct(ctx context.Context, product *providers.ProductPushRequest) (*providers.ProductPushResponse, error) {
	productBody := map[string]interface{}{
		"title":       product.Name,
		"description": product.Description,
		"category_id": product.CategoryID,
		"brand_id":    "",
		"images": func() []map[string]string {
			images := make([]map[string]string, len(product.Images))
			for i, img := range product.Images {
				images[i] = map[string]string{"id": img}
			}
			return images
		}(),
		"skus": []map[string]interface{}{
			{
				"seller_sku":       product.SKU,
				"original_price":   fmt.Sprintf("%.2f", product.OriginalPrice),
				"sales_attributes": []interface{}{},
				"stock_infos": []map[string]interface{}{
					{
						"available_stock": product.Stock,
						"warehouse_id":    "",
					},
				},
			},
		},
		"package_weight": fmt.Sprintf("%.2f", product.Weight/1000), // Convert g to kg
	}

	// Add dimensions if provided
	if product.Dimensions != nil {
		productBody["package_dimensions"] = map[string]interface{}{
			"length": fmt.Sprintf("%.0f", product.Dimensions.Length),
			"width":  fmt.Sprintf("%.0f", product.Dimensions.Width),
			"height": fmt.Sprintf("%.0f", product.Dimensions.Height),
			"unit":   "CM",
		}
	}

	req := &Request{
		Method:   http.MethodPost,
		Path:     CreateProductPath,
		Body:     productBody,
		NeedAuth: true,
	}

	var resp struct {
		BaseResponse
		Data struct {
			ProductID string `json:"product_id"`
			SKUs      []struct {
				ID        string `json:"id"`
				SellerSKU string `json:"seller_sku"`
			} `json:"skus"`
		} `json:"data"`
	}

	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to push product: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	externalSKU := ""
	if len(resp.Data.SKUs) > 0 {
		externalSKU = resp.Data.SKUs[0].ID
	}

	return &providers.ProductPushResponse{
		ExternalProductID: resp.Data.ProductID,
		ExternalSKU:       externalSKU,
		Status:            "created",
	}, nil
}

// UpdateProduct updates an existing product on TikTok Shop
func (p *ProductProvider) UpdateProduct(ctx context.Context, externalID string, product *providers.ProductUpdateRequest) error {
	updateBody := map[string]interface{}{
		"product_id": externalID,
	}

	if product.Name != "" {
		updateBody["title"] = product.Name
	}
	if product.Description != "" {
		updateBody["description"] = product.Description
	}

	req := &Request{
		Method:   http.MethodPut,
		Path:     UpdateProductPath + "/" + externalID,
		Body:     updateBody,
		NeedAuth: true,
	}

	var resp BaseResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	if resp.HasError() {
		return fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	return nil
}

// DeleteProduct deletes a product from TikTok Shop
func (p *ProductProvider) DeleteProduct(ctx context.Context, externalID string) error {
	req := &Request{
		Method:   http.MethodDelete,
		Path:     DeleteProductPath + "/" + externalID,
		NeedAuth: true,
	}

	var resp BaseResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if resp.HasError() {
		return fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	return nil
}

// UpdateInventory updates stock for products
func (p *ProductProvider) UpdateInventory(ctx context.Context, updates []providers.InventoryUpdate) error {
	stockUpdates := make([]map[string]interface{}, len(updates))
	for i, update := range updates {
		stockUpdates[i] = map[string]interface{}{
			"product_id": update.ExternalProductID,
			"sku_id":     update.ExternalSKU,
			"stock_infos": []map[string]interface{}{
				{
					"available_stock": update.Quantity,
				},
			},
		}
	}

	req := &Request{
		Method: http.MethodPut,
		Path:   UpdateInventoryPath,
		Body: map[string]interface{}{
			"skus": stockUpdates,
		},
		NeedAuth: true,
	}

	var resp BaseResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	if resp.HasError() {
		return fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	return nil
}

// GetInventory fetches inventory levels for products
func (p *ProductProvider) GetInventory(ctx context.Context, externalProductIDs []string) ([]providers.InventoryItem, error) {
	req := &Request{
		Method: http.MethodPost,
		Path:   GetProductsPath,
		Body: map[string]interface{}{
			"product_ids": externalProductIDs,
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
		return nil, fmt.Errorf("failed to get inventory: %w", err)
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
