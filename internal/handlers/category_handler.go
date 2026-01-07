package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/niaga-platform/service-marketplace/internal/models"
	"github.com/niaga-platform/service-marketplace/internal/services"
)

// CategoryHandler handles category mapping API requests
type CategoryHandler struct {
	service *services.ProductSyncService
	logger  *zap.Logger
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(service *services.ProductSyncService, logger *zap.Logger) *CategoryHandler {
	return &CategoryHandler{
		service: service,
		logger:  logger,
	}
}

// GetExternalCategories fetches categories from the marketplace
// GET /api/v1/admin/marketplace/connections/:id/categories/external
func (h *CategoryHandler) GetExternalCategories(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	categories, err := h.service.GetExternalCategories(c.Request.Context(), connectionID)
	if err != nil {
		h.logger.Error("Failed to get external categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"total":      len(categories),
	})
}

// GetCategoryMappings lists category mappings for a connection
// GET /api/v1/admin/marketplace/connections/:id/categories
func (h *CategoryHandler) GetCategoryMappings(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	mappings, err := h.service.GetCategoryMappings(c.Request.Context(), connectionID)
	if err != nil {
		h.logger.Error("Failed to get category mappings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mappings": mappings,
		"total":    len(mappings),
	})
}

// CreateCategoryMappingRequest represents the request to create a mapping
type CreateCategoryMappingRequest struct {
	InternalCategoryID   string `json:"internal_category_id" binding:"required"`
	ExternalCategoryID   string `json:"external_category_id" binding:"required"`
	ExternalCategoryName string `json:"external_category_name"`
}

// CreateCategoryMapping creates a new category mapping
// POST /api/v1/admin/marketplace/connections/:id/categories
func (h *CategoryHandler) CreateCategoryMapping(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var req CreateCategoryMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	internalCatID, err := uuid.Parse(req.InternalCategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid internal category ID"})
		return
	}

	mapping, err := h.service.CreateCategoryMapping(c.Request.Context(), connectionID, &models.CreateCategoryMappingRequest{
		InternalCategoryID:   internalCatID,
		ExternalCategoryID:   req.ExternalCategoryID,
		ExternalCategoryName: req.ExternalCategoryName,
	})
	if err != nil {
		h.logger.Error("Failed to create category mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Category mapping created",
		"mapping": mapping,
	})
}

// DeleteCategoryMapping deletes a category mapping
// DELETE /api/v1/admin/marketplace/connections/:id/categories/:mapping_id
func (h *CategoryHandler) DeleteCategoryMapping(c *gin.Context) {
	mappingID, err := uuid.Parse(c.Param("mapping_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	if err := h.service.DeleteCategoryMapping(c.Request.Context(), mappingID); err != nil {
		h.logger.Error("Failed to delete category mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category mapping deleted"})
}
