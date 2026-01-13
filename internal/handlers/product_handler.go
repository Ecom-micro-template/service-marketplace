package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Ecom-micro-template/service-marketplace/internal/domain"
	"github.com/Ecom-micro-template/service-marketplace/internal/application"
)

// ProductHandler handles product sync API requests
type ProductHandler struct {
	service *services.ProductSyncService
	logger  *zap.Logger
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(service *services.ProductSyncService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

// GetMappedProducts lists synced products for a connection
// GET /api/v1/admin/marketplace/connections/:id/products
func (h *ProductHandler) GetMappedProducts(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	filter := &models.ProductMappingFilter{
		SyncStatus: c.Query("status"),
		Page:       1,
		PageSize:   20,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			filter.PageSize = pageSize
		}
	}

	mappings, total, err := h.service.GetMappedProducts(c.Request.Context(), connectionID, filter)
	if err != nil {
		h.logger.Error("Failed to get mapped products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mappings": mappings,
		"total":    total,
		"page":     filter.Page,
		"pageSize": filter.PageSize,
	})
}

// PushProductsRequest represents the request to push products
// Empty product_ids means "push all active products"
type PushProductsRequest struct {
	ProductIDs []string `json:"product_ids"` // Optional - empty means push all
}

// PushProducts pushes products to a marketplace
// POST /api/v1/admin/marketplace/connections/:id/products/push
// If product_ids is empty, pushes all active products from catalog
func (h *ProductHandler) PushProducts(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var req PushProductsRequest
	_ = c.ShouldBindJSON(&req) // Ignore binding errors, empty is valid

	job, err := h.service.PushProducts(c.Request.Context(), connectionID, req.ProductIDs)
	if err != nil {
		h.logger.Error("Failed to push products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := "Product push job created"
	if len(req.ProductIDs) == 0 {
		message = "Push all products job created"
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": message,
		"job_id":  job.ID,
		"status":  job.Status,
	})
}

// UpdateProductMappingRequest represents the request to update a mapping
type UpdateProductMappingRequest struct {
	Status string `json:"status" binding:"required,oneof=synced pending error"`
}

// UpdateProductMapping updates a product mapping
// PUT /api/v1/admin/marketplace/connections/:id/products/:mapping_id
func (h *ProductHandler) UpdateProductMapping(c *gin.Context) {
	mappingID, err := uuid.Parse(c.Param("mapping_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	var req UpdateProductMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.UpdateProductMapping(c.Request.Context(), mappingID, req.Status); err != nil {
		h.logger.Error("Failed to update mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mapping updated"})
}

// DeleteProductMapping deletes a product mapping
// DELETE /api/v1/admin/marketplace/connections/:id/products/:mapping_id
func (h *ProductHandler) DeleteProductMapping(c *gin.Context) {
	mappingID, err := uuid.Parse(c.Param("mapping_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	if err := h.service.DeleteProductMapping(c.Request.Context(), mappingID); err != nil {
		h.logger.Error("Failed to delete mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mapping deleted"})
}

// ImportProducts imports products from a marketplace
// POST /api/v1/admin/marketplace/connections/:id/products/import
func (h *ProductHandler) ImportProducts(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	count, err := h.service.ImportProducts(c.Request.Context(), connectionID)
	if err != nil {
		h.logger.Error("Failed to import products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Products imported successfully",
		"products_imported": count,
	})
}

// GetImportedProducts retrieves imported products for a connection
// GET /api/v1/admin/marketplace/connections/:id/products/imported
func (h *ProductHandler) GetImportedProducts(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	filter := &models.ImportedProductFilter{
		Page:     1,
		PageSize: 20,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			filter.PageSize = pageSize
		}
	}
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}
	if isMapped := c.Query("is_mapped"); isMapped != "" {
		mapped := isMapped == "true"
		filter.IsMapped = &mapped
	}

	products, total, err := h.service.GetImportedProducts(c.Request.Context(), connectionID, filter)
	if err != nil {
		h.logger.Error("Failed to get imported products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get imported products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products":  products,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// CreateManualMappingRequest represents the request to create a manual mapping
type CreateManualMappingRequest struct {
	ImportedProductID string `json:"imported_product_id" binding:"required"`
	InternalProductID string `json:"internal_product_id" binding:"required"`
}

// CreateManualMapping creates a manual mapping between an imported product and an internal product
// POST /api/v1/admin/marketplace/connections/:id/products/map
func (h *ProductHandler) CreateManualMapping(c *gin.Context) {
	connectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var req CreateManualMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	importedProductID, err := uuid.Parse(req.ImportedProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid imported product ID"})
		return
	}

	internalProductID, err := uuid.Parse(req.InternalProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid internal product ID"})
		return
	}

	mapping, err := h.service.CreateManualMapping(c.Request.Context(), connectionID, importedProductID, internalProductID)
	if err != nil {
		h.logger.Error("Failed to create manual mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Mapping created successfully",
		"mapping": mapping,
	})
}

// DeleteManualMapping deletes a manual mapping
// DELETE /api/v1/admin/marketplace/connections/:id/products/map/:mapping_id
func (h *ProductHandler) DeleteManualMapping(c *gin.Context) {
	mappingID, err := uuid.Parse(c.Param("mapping_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	if err := h.service.DeleteManualMapping(c.Request.Context(), mappingID); err != nil {
		h.logger.Error("Failed to delete manual mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mapping deleted"})
}
