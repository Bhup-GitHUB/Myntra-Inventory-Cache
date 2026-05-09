package events

import "time"

const (
	ExchangeInventoryEvents = "inventory.events"
	RoutingInventoryUpdated = "inventory.updated"
	QueueCacheInvalidator   = "inventory.cache_invalidator"
)

type InventoryUpdatedEvent struct {
	EventID   string    `json:"event_id"`
	ProductID int64     `json:"product_id"`
	NewQty    int       `json:"new_qty"`
	Version   int64     `json:"version"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}
