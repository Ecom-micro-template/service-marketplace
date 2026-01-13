package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Ecom-micro-template/service-marketplace/internal/application"
)

// WebhookHandler handles incoming webhooks from marketplaces
type WebhookHandler struct {
	orderService *services.OrderSyncService
	shopeeKey    string
	tiktokSecret string
	logger       *zap.Logger
}

// WebhookConfig holds configuration for webhook handlers
type WebhookConfig struct {
	ShopeePartnerKey string
	TikTokAppSecret  string
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(orderService *services.OrderSyncService, cfg *WebhookConfig, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		orderService: orderService,
		shopeeKey:    cfg.ShopeePartnerKey,
		tiktokSecret: cfg.TikTokAppSecret,
		logger:       logger,
	}
}

// HandleShopeeWebhook handles incoming Shopee webhooks
func (h *WebhookHandler) HandleShopeeWebhook(c *gin.Context) {
	h.logger.Info("Shopee webhook received",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("remote_addr", c.ClientIP()),
	)

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	h.logger.Info("Shopee webhook body received", zap.Int("body_length", len(body)))

	// Check for push_verification_code (Shopee test push)
	var testPush struct {
		PushVerificationCode string `json:"push_verification_code"`
	}
	if err := json.Unmarshal(body, &testPush); err == nil && testPush.PushVerificationCode != "" {
		h.logger.Info("Shopee test push received", zap.String("verification_code", testPush.PushVerificationCode))
		// Return the verification code as required by Shopee
		c.JSON(http.StatusOK, gin.H{"push_verification_code": testPush.PushVerificationCode})
		return
	}

	// Verify signature if key is configured
	if h.shopeeKey != "" {
		signature := c.GetHeader("Authorization")
		h.logger.Info("Shopee signature check",
			zap.String("authorization_header", signature),
			zap.Bool("has_key", h.shopeeKey != ""),
		)
		if signature != "" && !h.verifyShopeeSignature(body, signature) {
			h.logger.Warn("Invalid Shopee webhook signature",
				zap.String("received_signature", signature),
			)
			// Don't reject - just log warning for now
		}
	}

	// Parse event
	var event struct {
		Code      int   `json:"code"`
		ShopID    int64 `json:"shop_id"`
		Timestamp int64 `json:"timestamp"`
		Data      struct {
			OrderSN string `json:"ordersn"`
			Status  string `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		h.logger.Error("Failed to parse Shopee webhook", zap.Error(err), zap.String("body", string(body)))
		// Still return 200 to acknowledge receipt
		c.JSON(http.StatusOK, gin.H{"status": "received", "warning": "parse error"})
		return
	}

	h.logger.Info("Received Shopee webhook event",
		zap.Int("code", event.Code),
		zap.Int64("shop_id", event.ShopID),
		zap.String("order_sn", event.Data.OrderSN),
		zap.String("status", event.Data.Status),
	)

	// Process order event
	if event.Data.OrderSN != "" {
		go h.orderService.HandleShopeeOrderEvent(event.ShopID, event.Data.OrderSN, event.Data.Status)
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (h *WebhookHandler) verifyShopeeSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.shopeeKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

// HandleTikTokWebhook handles incoming TikTok webhooks
func (h *WebhookHandler) HandleTikTokWebhook(c *gin.Context) {
	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify signature if secret is configured
	if h.tiktokSecret != "" {
		signature := c.GetHeader("X-Tts-Signature")
		if !h.verifyTikTokSignature(body, signature) {
			h.logger.Warn("Invalid TikTok webhook signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Parse event
	var event struct {
		Type      string `json:"type"`
		ShopID    string `json:"shop_id"`
		Timestamp int64  `json:"timestamp"`
		Data      struct {
			OrderID     string `json:"order_id"`
			OrderStatus int    `json:"order_status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		h.logger.Error("Failed to parse TikTok webhook", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	h.logger.Info("Received TikTok webhook",
		zap.String("type", event.Type),
		zap.String("shop_id", event.ShopID),
		zap.String("order_id", event.Data.OrderID),
	)

	// Process order event
	if event.Type == "ORDER_STATUS_CHANGE" && event.Data.OrderID != "" {
		go h.orderService.HandleTikTokOrderEvent(event.ShopID, event.Data.OrderID, event.Data.OrderStatus)
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (h *WebhookHandler) verifyTikTokSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.tiktokSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}
