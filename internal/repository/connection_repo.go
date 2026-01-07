package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/niaga-platform/service-marketplace/internal/models"
	"gorm.io/gorm"
)

// ConnectionRepository handles database operations for connections
type ConnectionRepository struct {
	db *gorm.DB
}

// NewConnectionRepository creates a new ConnectionRepository
func NewConnectionRepository(db *gorm.DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

// Create creates a new connection
func (r *ConnectionRepository) Create(ctx context.Context, connection *models.Connection) error {
	return r.db.WithContext(ctx).Create(connection).Error
}

// GetByID retrieves a connection by ID
func (r *ConnectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Connection, error) {
	var connection models.Connection
	err := r.db.WithContext(ctx).First(&connection, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &connection, nil
}

// GetByPlatformAndShopID retrieves a connection by platform and shop ID
func (r *ConnectionRepository) GetByPlatformAndShopID(ctx context.Context, platform, shopID string) (*models.Connection, error) {
	var connection models.Connection
	err := r.db.WithContext(ctx).
		Where("platform = ? AND shop_id = ?", platform, shopID).
		First(&connection).Error
	if err != nil {
		return nil, err
	}
	return &connection, nil
}

// GetActiveConnections retrieves all active connections
func (r *ConnectionRepository) GetActiveConnections(ctx context.Context) ([]models.Connection, error) {
	var connections []models.Connection
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&connections).Error
	return connections, err
}

// GetActiveConnectionsByPlatform retrieves active connections by platform
func (r *ConnectionRepository) GetActiveConnectionsByPlatform(ctx context.Context, platform string) ([]models.Connection, error) {
	var connections []models.Connection
	err := r.db.WithContext(ctx).
		Where("platform = ? AND is_active = ?", platform, true).
		Order("created_at DESC").
		Find(&connections).Error
	return connections, err
}

// GetAll retrieves all connections
func (r *ConnectionRepository) GetAll(ctx context.Context) ([]models.Connection, error) {
	var connections []models.Connection
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&connections).Error
	return connections, err
}

// Update updates a connection
func (r *ConnectionRepository) Update(ctx context.Context, connection *models.Connection) error {
	return r.db.WithContext(ctx).Save(connection).Error
}

// UpdateTokens updates only the token-related fields
func (r *ConnectionRepository) UpdateTokens(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.Connection{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"access_token":     accessToken,
			"refresh_token":    refreshToken,
			"token_expires_at": expiresAt,
		}).Error
}

// Deactivate deactivates a connection
func (r *ConnectionRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Connection{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// Delete deletes a connection
func (r *ConnectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Connection{}, "id = ?", id).Error
}

// GetConnectionsNeedingTokenRefresh gets connections whose tokens are about to expire
func (r *ConnectionRepository) GetConnectionsNeedingTokenRefresh(ctx context.Context, withinMinutes int) ([]models.Connection, error) {
	var connections []models.Connection
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND token_expires_at <= NOW() + INTERVAL '1 minute' * ?", true, withinMinutes).
		Find(&connections).Error
	return connections, err
}
