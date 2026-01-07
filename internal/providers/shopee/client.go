package shopee

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	ProductionBaseURL = "https://partner.shopeemobile.com"
	SandboxBaseURL    = "https://partner.test-stable.shopeemobile.com"
)

// Client is the Shopee API client
type Client struct {
	partnerID   int64
	partnerKey  string
	baseURL     string
	httpClient  *http.Client
	logger      *zap.Logger
	accessToken string
	shopID      int64
}

// ClientConfig holds configuration for the Shopee client
type ClientConfig struct {
	PartnerID   string
	PartnerKey  string
	IsSandbox   bool
	RedirectURL string
	Logger      *zap.Logger
}

// NewClient creates a new Shopee API client
func NewClient(cfg *ClientConfig) (*Client, error) {
	partnerID, err := strconv.ParseInt(cfg.PartnerID, 10, 64)
	if err != nil && cfg.PartnerID != "" {
		return nil, fmt.Errorf("invalid partner ID: %w", err)
	}

	baseURL := ProductionBaseURL
	if cfg.IsSandbox {
		baseURL = SandboxBaseURL
	}

	return &Client{
		partnerID:  partnerID,
		partnerKey: cfg.PartnerKey,
		baseURL:    baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: cfg.Logger,
	}, nil
}

// SetTokens sets the access token and shop ID for authenticated requests
func (c *Client) SetTokens(accessToken string, shopID int64) {
	c.accessToken = accessToken
	c.shopID = shopID
}

// generateSign generates the HMAC-SHA256 signature for Shopee API
func (c *Client) generateSign(path string, timestamp int64) string {
	baseString := fmt.Sprintf("%d%s%d", c.partnerID, path, timestamp)
	h := hmac.New(sha256.New, []byte(c.partnerKey))
	h.Write([]byte(baseString))
	return hex.EncodeToString(h.Sum(nil))
}

// generateSignWithTokens generates signature for authenticated endpoints
func (c *Client) generateSignWithTokens(path string, timestamp int64, accessToken string, shopID int64) string {
	baseString := fmt.Sprintf("%d%s%d%s%d", c.partnerID, path, timestamp, accessToken, shopID)
	h := hmac.New(sha256.New, []byte(c.partnerKey))
	h.Write([]byte(baseString))
	return hex.EncodeToString(h.Sum(nil))
}

// Request represents a generic API request
type Request struct {
	Method   string
	Path     string
	Query    map[string]string
	Body     interface{}
	NeedAuth bool
}

// Do performs an HTTP request to the Shopee API
func (c *Client) Do(ctx context.Context, req *Request, result interface{}) error {
	timestamp := time.Now().Unix()

	// Build URL with common params
	url := c.baseURL + req.Path
	queryParams := []string{
		fmt.Sprintf("partner_id=%d", c.partnerID),
		fmt.Sprintf("timestamp=%d", timestamp),
	}

	var sign string
	if req.NeedAuth {
		sign = c.generateSignWithTokens(req.Path, timestamp, c.accessToken, c.shopID)
		queryParams = append(queryParams, fmt.Sprintf("access_token=%s", c.accessToken))
		queryParams = append(queryParams, fmt.Sprintf("shop_id=%d", c.shopID))
	} else {
		sign = c.generateSign(req.Path, timestamp)
	}
	queryParams = append(queryParams, fmt.Sprintf("sign=%s", sign))

	// Add custom query params
	for k, v := range req.Query {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", k, v))
	}

	// Sort query params for consistency
	sort.Strings(queryParams)
	url += "?" + strings.Join(queryParams, "&")

	// Build request body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log for debugging
	c.logger.Debug("Shopee API response",
		zap.String("path", req.Path),
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	// Parse response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// BaseResponse is the common response structure from Shopee API
type BaseResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Warning string `json:"warning,omitempty"`
}

// HasError checks if the response contains an error
func (r *BaseResponse) HasError() bool {
	return r.Error != "" && r.Error != "success"
}

// GetError returns the error message
func (r *BaseResponse) GetError() string {
	if r.Message != "" {
		return r.Message
	}
	return r.Error
}
