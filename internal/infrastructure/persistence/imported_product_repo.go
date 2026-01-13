package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/Ecom-micro-template/service-marketplace/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ImportedProductRepository handles database operations for imported products
type ImportedProductRepository struct {
	db *gorm.DB
}

// NewImportedProductRepository creates a new ImportedProductRepository
func NewImportedProductRepository(db *gorm.DB) *ImportedProductRepository {
	return &ImportedProductRepository{db: db}
}

// Upsert creates or updates an imported product
func (r *ImportedProductRepository) Upsert(ctx context.Context, product *domain.ImportedProduct) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "connection_id"}, {Name: "external_product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description", "price", "stock", "category_id", "status", "image_url", "external_sku", "updated_at"}),
	}).Create(product).Error
}

// UpsertBatch creates or updates multiple imported products in a batch
func (r *ImportedProductRepository) UpsertBatch(ctx context.Context, products []domain.ImportedProduct) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "connection_id"}, {Name: "external_product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description", "price", "stock", "category_id", "status", "image_url", "external_sku", "updated_at"}),
	}).CreateInBatches(products, 50).Error
}

// GetByID retrieves an imported product by ID
func (r *ImportedProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ImportedProduct, error) {
	var product domain.ImportedProduct
	err := r.db.WithContext(ctx).First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByConnectionID retrieves all imported products for a connection
func (r *ImportedProductRepository) GetByConnectionID(ctx context.Context, connectionID uuid.UUID, filter *domain.ImportedProductFilter) ([]domain.ImportedProduct, int64, error) {
	var products []domain.ImportedProduct
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.ImportedProduct{}).Where("connection_id = ?", connectionID)

	if filter != nil {
		if filter.IsMapped != nil {
			query = query.Where("is_mapped = ?", *filter.IsMapped)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Search != "" {
			query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
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
		Order("imported_at DESC").
		Find(&products).Error

	return products, total, err
}

// GetByExternalProductID retrieves an imported product by connection and external product ID
func (r *ImportedProductRepository) GetByExternalProductID(ctx context.Context, connectionID uuid.UUID, externalProductID string) (*domain.ImportedProduct, error) {
	var product domain.ImportedProduct
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND external_product_id = ?", connectionID, externalProductID).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// SetMapped marks an imported product as mapped to an internal product
func (r *ImportedProductRepository) SetMapped(ctx context.Context, id uuid.UUID, internalProductID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.ImportedProduct{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_mapped":            true,
			"mapped_to_product_id": internalProductID,
		}).Error
}

// SetUnmapped marks an imported product as unmapped
func (r *ImportedProductRepository) SetUnmapped(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.ImportedProduct{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_mapped":            false,
			"mapped_to_product_id": nil,
		}).Error
}

// Delete deletes an imported product
func (r *ImportedProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.ImportedProduct{}, "id = ?", id).Error
}

// DeleteByConnectionID deletes all imported products for a connection
func (r *ImportedProductRepository) DeleteByConnectionID(ctx context.Context, connectionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("connection_id = ?", connectionID).
		Delete(&domain.ImportedProduct{}).Error
}

// GetUnmappedCount returns count of unmapped products
func (r *ImportedProductRepository) GetUnmappedCount(ctx context.Context, connectionID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.ImportedProduct{}).
		Where("connection_id = ? AND is_mapped = false", connectionID).
		Count(&count).Error
	return count, err
}
