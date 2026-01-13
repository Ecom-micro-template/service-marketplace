package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-marketplace/internal/domain"
	"gorm.io/gorm"
)

// ProductMappingRepository handles database operations for product mappings
type ProductMappingRepository struct {
	db *gorm.DB
}

// NewProductMappingRepository creates a new ProductMappingRepository
func NewProductMappingRepository(db *gorm.DB) *ProductMappingRepository {
	return &ProductMappingRepository{db: db}
}

// Create creates a new product mapping
func (r *ProductMappingRepository) Create(ctx context.Context, mapping *domain.ProductMapping) error {
	return r.db.WithContext(ctx).Create(mapping).Error
}

// CreateBatch creates multiple product mappings in a batch
func (r *ProductMappingRepository) CreateBatch(ctx context.Context, mappings []domain.ProductMapping) error {
	return r.db.WithContext(ctx).CreateInBatches(mappings, 100).Error
}

// GetByID retrieves a product mapping by ID
func (r *ProductMappingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProductMapping, error) {
	var mapping domain.ProductMapping
	err := r.db.WithContext(ctx).First(&mapping, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

// GetByConnectionAndInternalProduct retrieves a mapping by connection and internal product ID
func (r *ProductMappingRepository) GetByConnectionAndInternalProduct(ctx context.Context, connectionID, internalProductID uuid.UUID) (*domain.ProductMapping, error) {
	var mapping domain.ProductMapping
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND internal_product_id = ?", connectionID, internalProductID).
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

// GetByConnectionAndExternalProduct retrieves a mapping by connection and external product ID
func (r *ProductMappingRepository) GetByConnectionAndExternalProduct(ctx context.Context, connectionID uuid.UUID, externalProductID string) (*domain.ProductMapping, error) {
	var mapping domain.ProductMapping
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND external_product_id = ?", connectionID, externalProductID).
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

// GetByConnectionID retrieves all mappings for a connection
func (r *ProductMappingRepository) GetByConnectionID(ctx context.Context, connectionID uuid.UUID, filter *domain.ProductMappingFilter) ([]domain.ProductMapping, int64, error) {
	var mappings []domain.ProductMapping
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.ProductMapping{}).Where("connection_id = ?", connectionID)

	if filter != nil {
		if filter.SyncStatus != "" {
			query = query.Where("sync_status = ?", filter.SyncStatus)
		}
		if filter.InternalProductID != nil {
			query = query.Where("internal_product_id = ?", *filter.InternalProductID)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := 1
	pageSize := 20
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.PageSize > 0 {
			pageSize = filter.PageSize
		}
	}
	offset := (page - 1) * pageSize

	err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&mappings).Error

	return mappings, total, err
}

// GetByInternalProductID retrieves all mappings for an internal product across all connections
func (r *ProductMappingRepository) GetByInternalProductID(ctx context.Context, internalProductID uuid.UUID) ([]domain.ProductMapping, error) {
	var mappings []domain.ProductMapping
	err := r.db.WithContext(ctx).
		Preload("Connection").
		Where("internal_product_id = ?", internalProductID).
		Find(&mappings).Error
	return mappings, err
}

// Update updates a product mapping
func (r *ProductMappingRepository) Update(ctx context.Context, mapping *domain.ProductMapping) error {
	return r.db.WithContext(ctx).Save(mapping).Error
}

// UpdateSyncStatus updates the sync status of a mapping
func (r *ProductMappingRepository) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status, errorMessage string) error {
	updates := map[string]interface{}{
		"sync_status": status,
		"sync_error":  errorMessage,
	}
	if status == domain.SyncStatusSynced {
		updates["last_synced_at"] = gorm.Expr("NOW()")
		updates["sync_error"] = ""
	}
	return r.db.WithContext(ctx).
		Model(&domain.ProductMapping{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete deletes a product mapping
func (r *ProductMappingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.ProductMapping{}, "id = ?", id).Error
}

// DeleteByConnectionID deletes all mappings for a connection
func (r *ProductMappingRepository) DeleteByConnectionID(ctx context.Context, connectionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("connection_id = ?", connectionID).
		Delete(&domain.ProductMapping{}).Error
}

// GetPendingMappings retrieves mappings that need syncing
func (r *ProductMappingRepository) GetPendingMappings(ctx context.Context, connectionID uuid.UUID, limit int) ([]domain.ProductMapping, error) {
	var mappings []domain.ProductMapping
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND sync_status = ?", connectionID, domain.SyncStatusPending).
		Limit(limit).
		Find(&mappings).Error
	return mappings, err
}

// GetErrorMappings retrieves mappings that have sync errors
func (r *ProductMappingRepository) GetErrorMappings(ctx context.Context, connectionID uuid.UUID) ([]domain.ProductMapping, error) {
	var mappings []domain.ProductMapping
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND sync_status = ?", connectionID, domain.SyncStatusError).
		Find(&mappings).Error
	return mappings, err
}
