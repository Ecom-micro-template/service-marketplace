package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-marketplace/internal/domain"
	"gorm.io/gorm"
)

// CategoryMappingRepository handles database operations for category mappings
type CategoryMappingRepository struct {
	db *gorm.DB
}

// NewCategoryMappingRepository creates a new CategoryMappingRepository
func NewCategoryMappingRepository(db *gorm.DB) *CategoryMappingRepository {
	return &CategoryMappingRepository{db: db}
}

// Create creates a new category mapping
func (r *CategoryMappingRepository) Create(ctx context.Context, mapping *domain.CategoryMapping) error {
	return r.db.WithContext(ctx).Create(mapping).Error
}

// GetByID retrieves a category mapping by ID
func (r *CategoryMappingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CategoryMapping, error) {
	var mapping domain.CategoryMapping
	err := r.db.WithContext(ctx).First(&mapping, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

// GetByConnectionID retrieves all mappings for a connection
func (r *CategoryMappingRepository) GetByConnectionID(ctx context.Context, connectionID uuid.UUID) ([]domain.CategoryMapping, error) {
	var mappings []domain.CategoryMapping
	err := r.db.WithContext(ctx).
		Where("connection_id = ?", connectionID).
		Order("created_at DESC").
		Find(&mappings).Error
	return mappings, err
}

// GetByConnectionAndInternalCategory retrieves a mapping by connection and internal category
func (r *CategoryMappingRepository) GetByConnectionAndInternalCategory(ctx context.Context, connectionID, internalCategoryID uuid.UUID) (*domain.CategoryMapping, error) {
	var mapping domain.CategoryMapping
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND internal_category_id = ?", connectionID, internalCategoryID).
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

// GetByConnectionAndExternalCategory retrieves a mapping by connection and external category
func (r *CategoryMappingRepository) GetByConnectionAndExternalCategory(ctx context.Context, connectionID uuid.UUID, externalCategoryID string) (*domain.CategoryMapping, error) {
	var mapping domain.CategoryMapping
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND external_category_id = ?", connectionID, externalCategoryID).
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

// Update updates a category mapping
func (r *CategoryMappingRepository) Update(ctx context.Context, mapping *domain.CategoryMapping) error {
	return r.db.WithContext(ctx).Save(mapping).Error
}

// Delete deletes a category mapping
func (r *CategoryMappingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CategoryMapping{}, "id = ?", id).Error
}

// DeleteByConnectionID deletes all mappings for a connection
func (r *CategoryMappingRepository) DeleteByConnectionID(ctx context.Context, connectionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("connection_id = ?", connectionID).
		Delete(&domain.CategoryMapping{}).Error
}
