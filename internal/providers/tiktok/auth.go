package tiktok

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/niaga-platform/service-marketplace/internal/providers"
)

const (
	AuthURL          = "https://auth.tiktok-shops.com/oauth/authorize"
	TokenPath        = "/api/v2/token/get"
	RefreshTokenPath = "/api/v2/token/refresh"
	ShopInfoPath     = "/api/v2/seller/shop"
)

// AuthProvider implements OAuth methods for TikTok Shop
type AuthProvider struct {
	client      *Client
	redirectURL string
}

// NewAuthProvider creates a new TikTok auth provider
func NewAuthProvider(client *Client, redirectURL string) *AuthProvider {
	return &AuthProvider{
		client:      client,
		redirectURL: redirectURL,
	}
}

// GetPlatform returns the platform name
func (p *AuthProvider) GetPlatform() string {
	return "tiktok"
}

// GetAuthURL generates the OAuth authorization URL
func (p *AuthProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("app_key", p.client.appKey)
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", AuthURL, params.Encode())
}

// TokenResponse represents the token response from TikTok
type TokenResponse struct {
	BaseResponse
	Data struct {
		AccessToken          string `json:"access_token"`
		AccessTokenExpireIn  int64  `json:"access_token_expire_in"`
		RefreshToken         string `json:"refresh_token"`
		RefreshTokenExpireIn int64  `json:"refresh_token_expire_in"`
		OpenID               string `json:"open_id"`
		SellerName           string `json:"seller_name"`
		SellerBaseRegion     string `json:"seller_base_region"`
	} `json:"data"`
}

// ExchangeCode exchanges the authorization code for access/refresh tokens
func (p *AuthProvider) ExchangeCode(ctx context.Context, code string) (*providers.TokenResponse, error) {
	req := &Request{
		Method: http.MethodGet,
		Path:   TokenPath,
		Query: map[string]string{
			"auth_code":  code,
			"grant_type": "authorized_code",
		},
		NeedAuth: false,
	}

	var resp TokenResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	return &providers.TokenResponse{
		AccessToken:  resp.Data.AccessToken,
		RefreshToken: resp.Data.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.Data.AccessTokenExpireIn) * time.Second),
		ShopID:       resp.Data.OpenID,
		ShopName:     resp.Data.SellerName,
	}, nil
}

// RefreshToken refreshes an expired access token
func (p *AuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*providers.TokenResponse, error) {
	req := &Request{
		Method: http.MethodGet,
		Path:   RefreshTokenPath,
		Query: map[string]string{
			"refresh_token": refreshToken,
			"grant_type":    "refresh_token",
		},
		NeedAuth: false,
	}

	var resp TokenResponse
	if err := p.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if resp.HasError() {
		return nil, fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	return &providers.TokenResponse{
		AccessToken:  resp.Data.AccessToken,
		RefreshToken: resp.Data.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.Data.AccessTokenExpireIn) * time.Second),
		ShopID:       resp.Data.OpenID,
		ShopName:     resp.Data.SellerName,
	}, nil
}

// ShopInfoResponse represents shop info from TikTok
type ShopInfoResponse struct {
	BaseResponse
	Data struct {
		Shops []struct {
			ShopID   string `json:"shop_id"`
			ShopName string `json:"shop_name"`
			Region   string `json:"region"`
			Status   int    `json:"status"`
		} `json:"shops"`
	} `json:"data"`
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
		return nil, fmt.Errorf("tiktok error: %s", resp.GetError())
	}

	if len(resp.Data.Shops) == 0 {
		return nil, fmt.Errorf("no shops found")
	}

	shop := resp.Data.Shops[0]
	status := "active"
	if shop.Status != 1 {
		status = "inactive"
	}

	return &providers.ShopInfo{
		ShopID:   shop.ShopID,
		ShopName: shop.ShopName,
		Status:   status,
		Region:   shop.Region,
	}, nil
}
