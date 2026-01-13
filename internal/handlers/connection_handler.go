package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Ecom-micro-template/service-marketplace/internal/services"
)

// ConnectionHandler handles marketplace connection API requests
type ConnectionHandler struct {
	service *services.ConnectionService
	logger  *zap.Logger
}

// NewConnectionHandler creates a new ConnectionHandler
func NewConnectionHandler(service *services.ConnectionService, logger *zap.Logger) *ConnectionHandler {
	return &ConnectionHandler{
		service: service,
		logger:  logger,
	}
}

// GetConnections lists all marketplace connections
// GET /api/v1/admin/marketplace/connections
func (h *ConnectionHandler) GetConnections(c *gin.Context) {
	connections, err := h.service.GetAllConnections(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get connections", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get connections",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connections": connections,
		"total":       len(connections),
	})
}

// GetActiveConnections lists only active connections
// GET /api/v1/admin/marketplace/connections/active
func (h *ConnectionHandler) GetActiveConnections(c *gin.Context) {
	connections, err := h.service.GetActiveConnections(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get active connections", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get connections",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connections": connections,
		"total":       len(connections),
	})
}

// GetConnection retrieves a single connection by ID
// GET /api/v1/admin/marketplace/connections/:id
func (h *ConnectionHandler) GetConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid connection ID",
			"message": "ID must be a valid UUID",
		})
		return
	}

	connection, err := h.service.GetConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Connection not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connection": connection,
	})
}

// GetAuthURLRequest represents the request body for getting auth URL
type GetAuthURLRequest struct {
	State string `json:"state"` // Optional custom state
}

// GetAuthURL generates OAuth authorization URL for a platform
// POST /api/v1/admin/marketplace/:platform/auth-url
func (h *ConnectionHandler) GetAuthURL(c *gin.Context) {
	platform := c.Param("platform")

	if platform != "shopee" && platform != "tiktok" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid platform",
			"message": "Platform must be 'shopee' or 'tiktok'",
		})
		return
	}

	authURL, state, err := h.service.GetAuthURL(c.Request.Context(), platform)
	if err != nil {
		h.logger.Error("Failed to get auth URL", zap.String("platform", platform), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate auth URL",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
		"platform": platform,
	})
}

// HandleShopeeCallback handles Shopee OAuth callback
// GET /api/v1/admin/marketplace/shopee/callback
func (h *ConnectionHandler) HandleShopeeCallback(c *gin.Context) {
	code := c.Query("code")
	shopIDStr := c.Query("shop_id")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing authorization code",
			"message": "The 'code' parameter is required",
		})
		return
	}

	if shopIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing shop ID",
			"message": "The 'shop_id' parameter is required",
		})
		return
	}

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid shop ID",
			"message": "Shop ID must be a number",
		})
		return
	}

	connection, err := h.service.HandleShopeeCallback(c.Request.Context(), code, shopID)
	if err != nil {
		h.logger.Error("Failed to handle Shopee callback", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to connect Shopee shop",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Successfully connected Shopee shop",
		"connection": connection,
	})
}

// HandleTikTokCallback handles TikTok OAuth callback
// GET /api/v1/admin/marketplace/tiktok/callback
func (h *ConnectionHandler) HandleTikTokCallback(c *gin.Context) {
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing authorization code",
			"message": "The 'code' parameter is required",
		})
		return
	}

	connection, err := h.service.HandleTikTokCallback(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("Failed to handle TikTok callback", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to connect TikTok shop",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Successfully connected TikTok shop",
		"connection": connection,
	})
}

// Disconnect deactivates a marketplace connection
// DELETE /api/v1/admin/marketplace/connections/:id
func (h *ConnectionHandler) Disconnect(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid connection ID",
			"message": "ID must be a valid UUID",
		})
		return
	}

	if err := h.service.Disconnect(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to disconnect", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to disconnect",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully disconnected",
	})
}

// RefreshToken refreshes the access token for a connection
// POST /api/v1/admin/marketplace/connections/:id/refresh
func (h *ConnectionHandler) RefreshToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid connection ID",
			"message": "ID must be a valid UUID",
		})
		return
	}

	if err := h.service.RefreshConnectionToken(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to refresh token", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to refresh token",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
	})
}
