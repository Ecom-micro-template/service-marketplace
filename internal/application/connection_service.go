package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Ecom-micro-template/service-marketplace/internal/domain"
	"github.com/Ecom-micro-template/service-marketplace/internal/providers"
	"github.com/Ecom-micro-template/service-marketplace/internal/providers/shopee"
	"github.com/Ecom-micro-template/service-marketplace/internal/providers/tiktok"
	"github.com/Ecom-micro-template/service-marketplace/internal/infrastructure/persistence"
	"github.com/Ecom-micro-template/service-marketplace/internal/utils"
)

var (
	ErrInvalidPlatform    = errors.New("invalid platform: must be 'shopee' or 'tiktok'")
	ErrConnectionNotFound = errors.New("connection not found")
	ErrConnectionExists   = errors.New("connection already exists for this shop")
	ErrEncryptionRequired = errors.New("encryption key is required")
)

// ConnectionService handles marketplace connection operations
type ConnectionService struct {
	repo         *repository.ConnectionRepository
	encryptor    *utils.Encryptor
	shopeeClient *shopee.Client
	shopeeAuth   *shopee.AuthProvider
	tiktokClient *tiktok.Client
	tiktokAuth   *tiktok.AuthProvider
	logger       *zap.Logger
}

// ConnectionServiceConfig holds configuration for ConnectionService
type ConnectionServiceConfig struct {
	EncryptionKey     string
	ShopeePartnerID   string
	ShopeePartnerKey  string
	ShopeeRedirectURL string
	ShopeeSandbox     bool
	TikTokAppKey      string
	TikTokAppSecret   string
	TikTokRedirectURL string
}

// NewConnectionService creates a new ConnectionService
func NewConnectionService(
	repo *repository.ConnectionRepository,
	cfg *ConnectionServiceConfig,
	logger *zap.Logger,
) (*ConnectionService, error) {
	var encryptor *utils.Encryptor
	var err error

	if cfg.EncryptionKey != "" {
		encryptor, err = utils.NewEncryptor(cfg.EncryptionKey)
		if err != nil {
			logger.Warn("Failed to initialize encryptor, tokens will not be encrypted", zap.Error(err))
		}
	}

	// Initialize Shopee client
	var shopeeClient *shopee.Client
	var shopeeAuth *shopee.AuthProvider
	if cfg.ShopeePartnerID != "" && cfg.ShopeePartnerKey != "" {
		shopeeClient, err = shopee.NewClient(&shopee.ClientConfig{
			PartnerID:   cfg.ShopeePartnerID,
			PartnerKey:  cfg.ShopeePartnerKey,
			IsSandbox:   cfg.ShopeeSandbox,
			RedirectURL: cfg.ShopeeRedirectURL,
			Logger:      logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Shopee client: %w", err)
		}
		shopeeAuth = shopee.NewAuthProvider(shopeeClient, cfg.ShopeeRedirectURL)
	}

	// Initialize TikTok client
	var tiktokClient *tiktok.Client
	var tiktokAuth *tiktok.AuthProvider
	if cfg.TikTokAppKey != "" && cfg.TikTokAppSecret != "" {
		tiktokClient = tiktok.NewClient(&tiktok.ClientConfig{
			AppKey:      cfg.TikTokAppKey,
			AppSecret:   cfg.TikTokAppSecret,
			RedirectURL: cfg.TikTokRedirectURL,
			Logger:      logger,
		})
		tiktokAuth = tiktok.NewAuthProvider(tiktokClient, cfg.TikTokRedirectURL)
	}

	return &ConnectionService{
		repo:         repo,
		encryptor:    encryptor,
		shopeeClient: shopeeClient,
		shopeeAuth:   shopeeAuth,
		tiktokClient: tiktokClient,
		tiktokAuth:   tiktokAuth,
		logger:       logger,
	}, nil
}

// GetAllConnections retrieves all marketplace connections
func (s *ConnectionService) GetAllConnections(ctx context.Context) ([]models.ConnectionResponse, error) {
	connections, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	responses := make([]models.ConnectionResponse, len(connections))
	for i, conn := range connections {
		responses[i] = *conn.ToResponse()
	}

	return responses, nil
}

// GetActiveConnections retrieves all active connections
func (s *ConnectionService) GetActiveConnections(ctx context.Context) ([]models.ConnectionResponse, error) {
	connections, err := s.repo.GetActiveConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	responses := make([]models.ConnectionResponse, len(connections))
	for i, conn := range connections {
		responses[i] = *conn.ToResponse()
	}

	return responses, nil
}

// GetConnection retrieves a connection by ID
func (s *ConnectionService) GetConnection(ctx context.Context, id uuid.UUID) (*models.ConnectionResponse, error) {
	conn, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrConnectionNotFound
	}
	return conn.ToResponse(), nil
}

// generateState generates a random state for OAuth CSRF protection
func (s *ConnectionService) generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetAuthURL generates the OAuth authorization URL for a platform
func (s *ConnectionService) GetAuthURL(ctx context.Context, platform string) (string, string, error) {
	randomState, err := s.generateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Prefix state with platform name for callback detection
	state := fmt.Sprintf("%s_%s", platform, randomState)

	var authURL string
	switch platform {
	case "shopee":
		if s.shopeeAuth == nil {
			return "", "", errors.New("Shopee integration not configured")
		}
		authURL = s.shopeeAuth.GetAuthURL(state)
	case "tiktok":
		if s.tiktokAuth == nil {
			return "", "", errors.New("TikTok integration not configured")
		}
		authURL = s.tiktokAuth.GetAuthURL(state)
	default:
		return "", "", ErrInvalidPlatform
	}

	return authURL, state, nil
}

// HandleShopeeCallback handles the OAuth callback from Shopee
func (s *ConnectionService) HandleShopeeCallback(ctx context.Context, code string, shopID int64) (*models.ConnectionResponse, error) {
	if s.shopeeAuth == nil {
		return nil, errors.New("Shopee integration not configured")
	}

	// Exchange code for tokens
	tokenResp, err := s.shopeeAuth.ExchangeCode(ctx, code, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Set tokens for shop info request
	s.shopeeClient.SetTokens(tokenResp.AccessToken, shopID)

	// Get shop info
	shopInfo, err := s.shopeeAuth.GetShopInfo(ctx)
	if err != nil {
		s.logger.Warn("Failed to get shop info, using default", zap.Error(err))
		shopInfo = &providers.ShopInfo{ShopID: tokenResp.ShopID, ShopName: "Shopee Shop"}
	}

	// Encrypt tokens
	accessToken := tokenResp.AccessToken
	refreshToken := tokenResp.RefreshToken
	if s.encryptor != nil {
		accessToken, _ = s.encryptor.Encrypt(tokenResp.AccessToken)
		refreshToken, _ = s.encryptor.Encrypt(tokenResp.RefreshToken)
	}

	// Check if connection already exists
	existing, _ := s.repo.GetByPlatformAndShopID(ctx, "shopee", tokenResp.ShopID)
	if existing != nil {
		// Update existing connection
		existing.AccessToken = accessToken
		existing.RefreshToken = refreshToken
		existing.TokenExpiresAt = &tokenResp.ExpiresAt
		existing.IsActive = true
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update connection: %w", err)
		}
		return existing.ToResponse(), nil
	}

	// Create new connection
	conn := &models.Connection{
		Platform:       "shopee",
		ShopID:         tokenResp.ShopID,
		ShopName:       shopInfo.ShopName,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		TokenExpiresAt: &tokenResp.ExpiresAt,
		IsActive:       true,
	}

	if err := s.repo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	return conn.ToResponse(), nil
}

// HandleTikTokCallback handles the OAuth callback from TikTok
func (s *ConnectionService) HandleTikTokCallback(ctx context.Context, code string) (*models.ConnectionResponse, error) {
	if s.tiktokAuth == nil {
		return nil, errors.New("TikTok integration not configured")
	}

	// Exchange code for tokens
	tokenResp, err := s.tiktokAuth.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Encrypt tokens
	accessToken := tokenResp.AccessToken
	refreshToken := tokenResp.RefreshToken
	if s.encryptor != nil {
		accessToken, _ = s.encryptor.Encrypt(tokenResp.AccessToken)
		refreshToken, _ = s.encryptor.Encrypt(tokenResp.RefreshToken)
	}

	// Check if connection already exists
	existing, _ := s.repo.GetByPlatformAndShopID(ctx, "tiktok", tokenResp.ShopID)
	if existing != nil {
		// Update existing connection
		existing.AccessToken = accessToken
		existing.RefreshToken = refreshToken
		existing.TokenExpiresAt = &tokenResp.ExpiresAt
		existing.ShopName = tokenResp.ShopName
		existing.IsActive = true
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update connection: %w", err)
		}
		return existing.ToResponse(), nil
	}

	// Create new connection
	conn := &models.Connection{
		Platform:       "tiktok",
		ShopID:         tokenResp.ShopID,
		ShopName:       tokenResp.ShopName,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		TokenExpiresAt: &tokenResp.ExpiresAt,
		IsActive:       true,
	}

	if err := s.repo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	return conn.ToResponse(), nil
}

// Disconnect deactivates a connection
func (s *ConnectionService) Disconnect(ctx context.Context, id uuid.UUID) error {
	conn, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrConnectionNotFound
	}

	conn.IsActive = false
	return s.repo.Update(ctx, conn)
}

// RefreshConnectionToken refreshes the access token for a connection
func (s *ConnectionService) RefreshConnectionToken(ctx context.Context, id uuid.UUID) error {
	conn, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrConnectionNotFound
	}

	// Decrypt refresh token
	refreshToken := conn.RefreshToken
	if s.encryptor != nil {
		refreshToken, err = s.encryptor.Decrypt(conn.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to decrypt refresh token: %w", err)
		}
	}

	var newTokens struct {
		AccessToken  string
		RefreshToken string
		ExpiresAt    time.Time
	}

	switch conn.Platform {
	case "shopee":
		if s.shopeeAuth == nil {
			return errors.New("Shopee integration not configured")
		}
		shopID, _ := strconv.ParseInt(conn.ShopID, 10, 64)
		resp, err := s.shopeeAuth.RefreshToken(ctx, refreshToken, shopID)
		if err != nil {
			return fmt.Errorf("failed to refresh Shopee token: %w", err)
		}
		newTokens.AccessToken = resp.AccessToken
		newTokens.RefreshToken = resp.RefreshToken
		newTokens.ExpiresAt = resp.ExpiresAt

	case "tiktok":
		if s.tiktokAuth == nil {
			return errors.New("TikTok integration not configured")
		}
		resp, err := s.tiktokAuth.RefreshToken(ctx, refreshToken)
		if err != nil {
			return fmt.Errorf("failed to refresh TikTok token: %w", err)
		}
		newTokens.AccessToken = resp.AccessToken
		newTokens.RefreshToken = resp.RefreshToken
		newTokens.ExpiresAt = resp.ExpiresAt

	default:
		return ErrInvalidPlatform
	}

	// Encrypt new tokens
	accessToken := newTokens.AccessToken
	refreshTokenNew := newTokens.RefreshToken
	if s.encryptor != nil {
		accessToken, _ = s.encryptor.Encrypt(newTokens.AccessToken)
		refreshTokenNew, _ = s.encryptor.Encrypt(newTokens.RefreshToken)
	}

	return s.repo.UpdateTokens(ctx, id, accessToken, refreshTokenNew, newTokens.ExpiresAt)
}
