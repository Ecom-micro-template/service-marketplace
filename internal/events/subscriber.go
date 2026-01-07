package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Event subjects
const (
	SubjectInventoryStockChanged = "inventory.stock.changed"
	SubjectMarketplaceSyncOK     = "marketplace.sync.completed"
	SubjectMarketplaceSyncFailed = "marketplace.sync.failed"
)

// StockChangedEvent represents an inventory change event
type StockChangedEvent struct {
	ProductID   uuid.UUID  `json:"product_id"`
	VariantID   *uuid.UUID `json:"variant_id,omitempty"`
	SKU         string     `json:"sku"`
	OldQuantity int        `json:"old_quantity"`
	NewQuantity int        `json:"new_quantity"`
	WarehouseID string     `json:"warehouse_id,omitempty"`
	Reason      string     `json:"reason"` // sale, adjustment, return, etc.
	Timestamp   time.Time  `json:"timestamp"`
}

// SyncCompletedEvent represents a successful sync event
type SyncCompletedEvent struct {
	ConnectionID uuid.UUID `json:"connection_id"`
	Platform     string    `json:"platform"`
	ProductID    uuid.UUID `json:"product_id"`
	SyncType     string    `json:"sync_type"` // inventory, product, order
	Timestamp    time.Time `json:"timestamp"`
}

// SyncFailedEvent represents a failed sync event
type SyncFailedEvent struct {
	ConnectionID uuid.UUID `json:"connection_id"`
	Platform     string    `json:"platform"`
	ProductID    uuid.UUID `json:"product_id"`
	SyncType     string    `json:"sync_type"`
	Error        string    `json:"error"`
	Timestamp    time.Time `json:"timestamp"`
}

// Subscriber handles NATS event subscriptions
type Subscriber struct {
	nc      *nats.Conn
	logger  *zap.Logger
	handler EventHandler
	subs    []*nats.Subscription
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	HandleStockChanged(event *StockChangedEvent) error
}

// NewSubscriber creates a new NATS subscriber
func NewSubscriber(nc *nats.Conn, handler EventHandler, logger *zap.Logger) *Subscriber {
	return &Subscriber{
		nc:      nc,
		logger:  logger,
		handler: handler,
		subs:    make([]*nats.Subscription, 0),
	}
}

// Start subscribes to all relevant events
func (s *Subscriber) Start() error {
	// Subscribe to inventory changes
	sub, err := s.nc.Subscribe(SubjectInventoryStockChanged, s.handleStockChanged)
	if err != nil {
		return err
	}
	s.subs = append(s.subs, sub)

	s.logger.Info("NATS subscriber started", zap.String("subject", SubjectInventoryStockChanged))
	return nil
}

// Stop unsubscribes from all events
func (s *Subscriber) Stop() {
	for _, sub := range s.subs {
		sub.Unsubscribe()
	}
	s.logger.Info("NATS subscriber stopped")
}

// handleStockChanged processes stock changed events
func (s *Subscriber) handleStockChanged(msg *nats.Msg) {
	var event StockChangedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("Failed to unmarshal stock changed event", zap.Error(err))
		return
	}

	s.logger.Info("Received stock changed event",
		zap.String("product_id", event.ProductID.String()),
		zap.Int("new_quantity", event.NewQuantity),
	)

	if err := s.handler.HandleStockChanged(&event); err != nil {
		s.logger.Error("Failed to handle stock changed event", zap.Error(err))
	}
}

// Publisher handles publishing events to NATS
type Publisher struct {
	nc     *nats.Conn
	logger *zap.Logger
}

// NewPublisher creates a new NATS publisher
func NewPublisher(nc *nats.Conn, logger *zap.Logger) *Publisher {
	return &Publisher{nc: nc, logger: logger}
}

// PublishSyncCompleted publishes a sync completed event
func (p *Publisher) PublishSyncCompleted(event *SyncCompletedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.nc.Publish(SubjectMarketplaceSyncOK, data)
}

// PublishSyncFailed publishes a sync failed event
func (p *Publisher) PublishSyncFailed(event *SyncFailedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.nc.Publish(SubjectMarketplaceSyncFailed, data)
}
