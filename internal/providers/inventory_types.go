package providers

// InventoryUpdateResult represents the result of an inventory update
type InventoryUpdateResult struct {
	ExternalProductID string `json:"external_product_id"`
	ExternalSKU       string `json:"external_sku,omitempty"`
	Success           bool   `json:"success"`
	Error             string `json:"error,omitempty"`
}
