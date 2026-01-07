package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niaga-platform/service-marketplace/internal/models"
	"gorm.io/gorm"
)

// MarketplaceOrderRepository handles database operations for marketplace orders
type MarketplaceOrderRepository struct {
	db *gorm.DB
}

// NewMarketplaceOrderRepository creates a new MarketplaceOrderRepository
func NewMarketplaceOrderRepository(db *gorm.DB) *MarketplaceOrderRepository {
	return &MarketplaceOrderRepository{db: db}
}

// Create creates a new marketplace order
func (r *MarketplaceOrderRepository) Create(ctx context.Context, order *models.MarketplaceOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetByID retrieves an order by ID
func (r *MarketplaceOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MarketplaceOrder, error) {
	var order models.MarketplaceOrder
	err := r.db.WithContext(ctx).
		Preload("Connection").
		First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByExternalOrderID retrieves an order by external order ID and connection
func (r *MarketplaceOrderRepository) GetByExternalOrderID(ctx context.Context, connectionID uuid.UUID, externalOrderID string) (*models.MarketplaceOrder, error) {
	var order models.MarketplaceOrder
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND external_order_id = ?", connectionID, externalOrderID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByInternalOrderID retrieves an order by internal order ID
func (r *MarketplaceOrderRepository) GetByInternalOrderID(ctx context.Context, internalOrderID uuid.UUID) (*models.MarketplaceOrder, error) {
	var order models.MarketplaceOrder
	err := r.db.WithContext(ctx).
		Preload("Connection").
		Where("internal_order_id = ?", internalOrderID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByConnectionID retrieves orders by connection with filters
func (r *MarketplaceOrderRepository) GetByConnectionID(ctx context.Context, connectionID uuid.UUID, filter *models.MarketplaceOrderFilter) ([]models.MarketplaceOrder, int64, error) {
	var orders []models.MarketplaceOrder
	var total int64

	query := r.db.WithContext(ctx).Model(&models.MarketplaceOrder{}).Where("connection_id = ?", connectionID)

	if filter != nil {
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.ExternalOrderID != "" {
			query = query.Where("external_order_id ILIKE ?", "%"+filter.ExternalOrderID+"%")
		}
		if filter.ImportedOnly != nil && *filter.ImportedOnly {
			query = query.Where("internal_order_id IS NOT NULL")
		}
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", *filter.EndDate)
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
		Find(&orders).Error

	return orders, total, err
}

// GetAllByPlatform retrieves orders by platform
func (r *MarketplaceOrderRepository) GetAllByPlatform(ctx context.Context, platform string, filter *models.MarketplaceOrderFilter) ([]models.MarketplaceOrder, int64, error) {
	var orders []models.MarketplaceOrder
	var total int64

	query := r.db.WithContext(ctx).Model(&models.MarketplaceOrder{}).
		Preload("Connection").
		Where("platform = ?", platform)

	if filter != nil {
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", *filter.EndDate)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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
		Find(&orders).Error

	return orders, total, err
}

// Update updates a marketplace order
func (r *MarketplaceOrderRepository) Update(ctx context.Context, order *models.MarketplaceOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// UpdateStatus updates the status of an order
func (r *MarketplaceOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.MarketplaceOrder{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// LinkToInternalOrder links a marketplace order to an internal order
func (r *MarketplaceOrderRepository) LinkToInternalOrder(ctx context.Context, id, internalOrderID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.MarketplaceOrder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"internal_order_id": internalOrderID,
			"synced_at":         now,
		}).Error
}

// Delete deletes a marketplace order
func (r *MarketplaceOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.MarketplaceOrder{}, "id = ?", id).Error
}

// GetUnimportedOrders retrieves orders that haven't been imported yet
func (r *MarketplaceOrderRepository) GetUnimportedOrders(ctx context.Context, connectionID uuid.UUID) ([]models.MarketplaceOrder, error) {
	var orders []models.MarketplaceOrder
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND internal_order_id IS NULL", connectionID).
		Order("created_at ASC").
		Find(&orders).Error
	return orders, err
}

// GetOrderStats retrieves order statistics for a connection
func (r *MarketplaceOrderRepository) GetOrderStats(ctx context.Context, connectionID uuid.UUID) (map[string]interface{}, error) {
	var stats struct {
		TotalOrders    int64   `json:"total_orders"`
		ImportedOrders int64   `json:"imported_orders"`
		PendingOrders  int64   `json:"pending_orders"`
		TotalRevenue   float64 `json:"total_revenue"`
	}

	r.db.WithContext(ctx).Model(&models.MarketplaceOrder{}).
		Where("connection_id = ?", connectionID).
		Count(&stats.TotalOrders)

	r.db.WithContext(ctx).Model(&models.MarketplaceOrder{}).
		Where("connection_id = ? AND internal_order_id IS NOT NULL", connectionID).
		Count(&stats.ImportedOrders)

	r.db.WithContext(ctx).Model(&models.MarketplaceOrder{}).
		Where("connection_id = ? AND internal_order_id IS NULL", connectionID).
		Count(&stats.PendingOrders)

	r.db.WithContext(ctx).Model(&models.MarketplaceOrder{}).
		Where("connection_id = ?", connectionID).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&stats.TotalRevenue)

	return map[string]interface{}{
		"total_orders":    stats.TotalOrders,
		"imported_orders": stats.ImportedOrders,
		"pending_orders":  stats.PendingOrders,
		"total_revenue":   stats.TotalRevenue,
	}, nil
}
