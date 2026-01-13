package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SyncJob represents a background sync job in the queue
type SyncJob struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConnectionID uuid.UUID      `gorm:"type:uuid" json:"connection_id"`
	JobType      string         `gorm:"type:varchar(50);not null" json:"job_type"` // product_push, inventory_sync, order_sync
	Payload      datatypes.JSON `gorm:"type:jsonb;not null" json:"payload"`
	Status       string         `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, processing, completed, failed
	Attempts     int            `gorm:"default:0" json:"attempts"`
	MaxAttempts  int            `gorm:"default:3" json:"max_attempts"`
	ErrorMessage string         `gorm:"type:text" json:"error_message,omitempty"`
	ScheduledAt  time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"scheduled_at"`
	StartedAt    *time.Time     `gorm:"type:timestamptz" json:"started_at"`
	CompletedAt  *time.Time     `gorm:"type:timestamptz" json:"completed_at"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Connection *Connection `gorm:"foreignKey:ConnectionID" json:"connection,omitempty"`
}

// TableName specifies the table name for SyncJob
func (SyncJob) TableName() string {
	return "marketplace.sync_jobs"
}

// Job type constants
const (
	JobTypeProductPush   = "product_push"
	JobTypeProductUpdate = "product_update"
	JobTypeInventorySync = "inventory_sync"
	JobTypeOrderSync     = "order_sync"
	JobTypeTokenRefresh  = "token_refresh"
)

// Job status constants
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

// ProductPushPayload represents the payload for a product push job
type ProductPushPayload struct {
	InternalProductIDs []uuid.UUID `json:"internal_product_ids"`
	CategoryMappingID  uuid.UUID   `json:"category_mapping_id"`
}

// InventorySyncPayload represents the payload for an inventory sync job
type InventorySyncPayload struct {
	InternalProductID uuid.UUID `json:"internal_product_id"`
	NewQuantity       int       `json:"new_quantity"`
	WarehouseID       string    `json:"warehouse_id,omitempty"`
}

// OrderSyncPayload represents the payload for an order sync job
type OrderSyncPayload struct {
	ExternalOrderID string `json:"external_order_id"`
	Action          string `json:"action"` // fetch, import, update_status
}

// SyncJobFilter represents filter options for sync jobs
type SyncJobFilter struct {
	ConnectionID *uuid.UUID `json:"connection_id"`
	JobType      string     `json:"job_type"`
	Status       string     `json:"status"`
	Page         int        `json:"page"`
	PageSize     int        `json:"page_size"`
}

// CreateSyncJobRequest represents a request to create a sync job
type CreateSyncJobRequest struct {
	ConnectionID uuid.UUID      `json:"connection_id" binding:"required"`
	JobType      string         `json:"job_type" binding:"required"`
	Payload      datatypes.JSON `json:"payload" binding:"required"`
	ScheduledAt  *time.Time     `json:"scheduled_at"`
}
