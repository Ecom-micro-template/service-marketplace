package shopee

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Ecom-micro-template/service-marketplace/internal/providers"
)

const (
	AuthPath         = "/api/v2/shop/auth_partner"
	TokenPath        = "/api/v2/auth/token/get"
	RefreshTokenPath = "/api/v2/auth/access_token/get"
	ShopInfoPath     = "/api/v2/shop/get_shop_info"
)

// AuthProvider implements OAuth methods for Shopee
type AuthProvider struct {
	client      *Client
	redirectURL string
}

// NewAuthProvider creates a new Shopee auth provider
func NewAuthProvider(client *Client, redirectURL string) *AuthProvider {
	return &AuthProvider{
		client:      client,
		redirectURL: redirectURL,
	}
}

// GetPlatform returns the platform name
func (p *AuthProvider) GetPlatform() string {
	return "shopee"
}

// GetAuthURL generates the OAuth authorization URL
func (p *AuthProvider) GetAuthURL(state string) string {
	timestamp := time.Now().Unix()
	path := AuthPath

	sign := p.client.generateSign(path, timestamp)

	// URL-encode the redirect URL
	encodedRedirect := url.QueryEscape(p.redirectURL)

	authURL := fmt.Sprintf("%s%s?partner_id=%d&timestamp=%d&sign=%s&redirect=%s",
		p.client.baseURL,
		path,
		p.client.partnerID,
		timestamp,
		sign,
		encodedRedirect,
	)

	// Append state if provided (for CSRF protection)
	if state != "" {
		authURL += "&state=" + url.QueryEscape(state)
	}

	return authURL
}

// TokenResponse represents the token response from Shopee
type TokenResponse struct {
	BaseResponse
	AccessToken    string  `json:"access_token"`
	RefreshToken   string  `json:"refresh_token"`
	ExpireIn       int64   `json:"expire_in"` // seconds until expiry
	ShopIDList     []int64 `json:"shop_id_list,omitempty"`
	MerchantIDList []int64 `json:"merchant_id_list,omitempty"`
}

// ExchangeCode exchanges the authorization code for access/refresh tokens
func (p *AuthProvider) ExchangeCode(ctx context.Context, code string, shopID int64) (*providers.TokenResponse, error) {
	req := &Request{
		Method: http.MethodPost,
		Path:   TokenPath,
		Body: map[string]interface{}{
			"code":       code,
			"shop_id":    shopID,
			"partner_id": p.client.partnerID,
		},
		NeedAuth: false,
	}

	var resp TokenResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("shopee error: %s", resp.GetError())
	}

	return &providers.TokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpireIn) * time.Second),
		ShopID:       fmt.Sprintf("%d", shopID),
	}, nil
}

// RefreshToken refreshes an expired access token
func (p *AuthProvider) RefreshToken(ctx context.Context, refreshToken string, shopID int64) (*providers.TokenResponse, error) {
	req := &Request{
		Method: http.MethodPost,
		Path:   RefreshTokenPath,
		Body: map[string]interface{}{
			"refresh_token": refreshToken,
			"shop_id":       shopID,
			"partner_id":    p.client.partnerID,
		},
		NeedAuth: false,
	}

	var resp TokenResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("shopee error: %s", resp.GetError())
	}

	return &providers.TokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpireIn) * time.Second),
		ShopID:       fmt.Sprintf("%d", shopID),
	}, nil
}

// ShopInfoResponse represents shop info from Shopee
type ShopInfoResponse struct {
	BaseResponse
	ShopName        string `json:"shop_name"`
	Region          string `json:"region"`
	Status          string `json:"status"`
	ShopDescription string `json:"shop_description"`
	ShopLogo        string `json:"shop_logo"`
}

// GetShopInfo fetches shop information
func (p *AuthProvider) GetShopInfo(ctx context.Context) (*providers.ShopInfo, error) {
	req := &Request{
		Method:   http.MethodGet,
		Path:     ShopInfoPath,
		NeedAuth: true,
	}

	var resp ShopInfoResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get shop info: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("shopee error: %s", resp.GetError())
	}

	return &providers.ShopInfo{
		ShopID:      fmt.Sprintf("%d", p.client.shopID),
		ShopName:    resp.ShopName,
		Status:      resp.Status,
		Region:      resp.Region,
		ShopLogo:    resp.ShopLogo,
		Description: resp.ShopDescription,
	}, nil
}
