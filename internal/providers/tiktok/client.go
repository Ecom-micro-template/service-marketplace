package tiktok

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
	"net/url"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	BaseURL = "https://open-api.tiktokglobalshop.com"
)

// Client is the TikTok Shop API client
type Client struct {
	appKey      string
	appSecret   string
	baseURL     string
	httpClient  *http.Client
	logger      *zap.Logger
	accessToken string
	shopID      string
}

// ClientConfig holds configuration for the TikTok client
type ClientConfig struct {
	AppKey      string
	AppSecret   string
	RedirectURL string
	Logger      *zap.Logger
}

// NewClient creates a new TikTok Shop API client
func NewClient(cfg *ClientConfig) *Client {
	return &Client{
		appKey:    cfg.AppKey,
		appSecret: cfg.AppSecret,
		baseURL:   BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: cfg.Logger,
	}
}

// SetTokens sets the access token and shop ID for authenticated requests
func (c *Client) SetTokens(accessToken, shopID string) {
	c.accessToken = accessToken
	c.shopID = shopID
}

// generateSign generates the HMAC-SHA256 signature for TikTok API
func (c *Client) generateSign(path string, timestamp int64, params map[string]string) string {
	// Collect all params except sign and access_token
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "access_token" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build sign string: secret + path + sorted params + secret
	var signBuilder strings.Builder
	signBuilder.WriteString(c.appSecret)
	signBuilder.WriteString(path)
	for _, k := range keys {
		signBuilder.WriteString(k)
		signBuilder.WriteString(params[k])
	}
	signBuilder.WriteString(c.appSecret)

	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(signBuilder.String()))
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

// Do performs an HTTP request to the TikTok API
func (c *Client) Do(ctx context.Context, req *Request, result interface{}) error {
	timestamp := time.Now().Unix()

	// Build query params
	params := map[string]string{
		"app_key":   c.appKey,
		"timestamp": fmt.Sprintf("%d", timestamp),
	}

	if req.NeedAuth && c.accessToken != "" {
		params["access_token"] = c.accessToken
	}
	if c.shopID != "" {
		params["shop_id"] = c.shopID
	}

	// Add custom query params
	for k, v := range req.Query {
		params[k] = v
	}

	// Generate signature
	sign := c.generateSign(req.Path, timestamp, params)
	params["sign"] = sign

	// Build URL
	u, err := url.Parse(c.baseURL + req.Path)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

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
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), bodyReader)
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
	c.logger.Debug("TikTok API response",
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

// BaseResponse is the common response structure from TikTok API
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HasError checks if the response contains an error
func (r *BaseResponse) HasError() bool {
	return r.Code != 0
}

// GetError returns the error message
func (r *BaseResponse) GetError() string {
	return r.Message
}
