package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Ecom-micro-template/service-marketplace/internal/providers"
	"github.com/Ecom-micro-template/service-marketplace/internal/application"
)

// InventoryHandler handles inventory sync API requests
type InventoryHandler struct {
	service *services.InventorySyncService
	logger  *zap.Logger
}

// NewInventoryHandler creates a new InventoryHandler
func NewInventoryHandler(service *services.InventorySyncService, logger *zap.Logger) *InventoryHandler {
	return &InventoryHandler{
		service: service,
		logger:  logger,
	}
}

// PushInventoryRequest represents the request to push inventory
type PushInventoryRequest struct {
	Updates []struct {
		ExternalProductID string `json:"external_product_id" binding:"required"`
		ExternalSKU       string `json:"external_sku"`
		Quantity          int    `json:"quantity" binding:"min=0"`
	} `json:"updates" binding:"required,min=1"`
}

// PushInventory manually pushes inventory updates
// POST /api/v1/admin/marketplace/connections/:id/inventory/push
func (h *InventoryHandler) PushInventory(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var req PushInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updates := make([]providers.InventoryUpdate, len(req.Updates))
	for i, u := range req.Updates {
		updates[i] = providers.InventoryUpdate{
			ExternalProductID: u.ExternalProductID,
			ExternalSKU:       u.ExternalSKU,
			Quantity:          u.Quantity,
		}
	}

	results, err := h.service.PushInventory(c.Request.Context(), connectionID, updates)
	if err != nil {
		h.logger.Error("Failed to push inventory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Inventory push completed",
		"success_count": successCount,
		"total_count":   len(results),
		"results":       results,
	})
}

// GetInventoryStatusRequest represents the request to get inventory status
type GetInventoryStatusRequest struct {
	ProductIDs []string `json:"product_ids" binding:"required,min=1"`
}

// GetInventoryStatus fetches inventory status from marketplace
// POST /api/v1/admin/marketplace/connections/:id/inventory/status
func (h *InventoryHandler) GetInventoryStatus(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var req GetInventoryStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	items, err := h.service.GetInventoryStatus(c.Request.Context(), connectionID, req.ProductIDs)
	if err != nil {
		h.logger.Error("Failed to get inventory status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inventory": items,
		"total":     len(items),
	})
}
