package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niaga-platform/service-marketplace/internal/models"
	"gorm.io/gorm"
)

// SyncJobRepository handles database operations for sync jobs
type SyncJobRepository struct {
	db *gorm.DB
}

// NewSyncJobRepository creates a new SyncJobRepository
func NewSyncJobRepository(db *gorm.DB) *SyncJobRepository {
	return &SyncJobRepository{db: db}
}

// Create creates a new sync job
func (r *SyncJobRepository) Create(ctx context.Context, job *models.SyncJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

// GetByID retrieves a sync job by ID
func (r *SyncJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SyncJob, error) {
	var job models.SyncJob
	err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetByConnectionID retrieves jobs for a connection with optional filters
func (r *SyncJobRepository) GetByConnectionID(ctx context.Context, connectionID uuid.UUID, filter *models.SyncJobFilter) ([]models.SyncJob, int64, error) {
	var jobs []models.SyncJob
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SyncJob{}).Where("connection_id = ?", connectionID)

	if filter != nil {
		if filter.JobType != "" {
			query = query.Where("job_type = ?", filter.JobType)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
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
		Find(&jobs).Error

	return jobs, total, err
}

// GetPendingJobs retrieves pending jobs for processing
func (r *SyncJobRepository) GetPendingJobs(ctx context.Context, limit int) ([]models.SyncJob, error) {
	var jobs []models.SyncJob
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ? AND attempts < max_attempts", models.JobStatusPending, now).
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

// GetFailedJobs retrieves failed jobs that can be retried
func (r *SyncJobRepository) GetFailedJobs(ctx context.Context, connectionID uuid.UUID) ([]models.SyncJob, error) {
	var jobs []models.SyncJob
	err := r.db.WithContext(ctx).
		Where("connection_id = ? AND status = ? AND attempts < max_attempts", connectionID, models.JobStatusFailed).
		Order("created_at DESC").
		Find(&jobs).Error
	return jobs, err
}

// Update updates a sync job
func (r *SyncJobRepository) Update(ctx context.Context, job *models.SyncJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

// MarkProcessing marks a job as processing
func (r *SyncJobRepository) MarkProcessing(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SyncJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     models.JobStatusProcessing,
			"started_at": now,
			"attempts":   gorm.Expr("attempts + 1"),
		}).Error
}

// MarkCompleted marks a job as completed
func (r *SyncJobRepository) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SyncJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.JobStatusCompleted,
			"completed_at": now,
		}).Error
}

// MarkFailed marks a job as failed with error message
func (r *SyncJobRepository) MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) error {
	return r.db.WithContext(ctx).
		Model(&models.SyncJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.JobStatusFailed,
			"error_message": errorMessage,
		}).Error
}

// Delete deletes a sync job
func (r *SyncJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.SyncJob{}, "id = ?", id).Error
}

// DeleteOldCompleted deletes completed jobs older than specified hours
func (r *SyncJobRepository) DeleteOldCompleted(ctx context.Context, olderThanHours int) error {
	cutoff := time.Now().Add(-time.Duration(olderThanHours) * time.Hour)
	return r.db.WithContext(ctx).
		Where("status = ? AND completed_at < ?", models.JobStatusCompleted, cutoff).
		Delete(&models.SyncJob{}).Error
}
