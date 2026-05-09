package checkout

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"inventory-cache-lab/internal/events"
	"inventory-cache-lab/internal/inventory"
)

type InventoryEventPublisher interface {
	PublishInventoryUpdated(ctx context.Context, event events.InventoryUpdatedEvent) error
}

type Service struct {
	db        *sql.DB
	publisher InventoryEventPublisher
	logger    *slog.Logger
}

func NewService(db *sql.DB, publisher InventoryEventPublisher, logger *slog.Logger) *Service {
	return &Service{db: db, publisher: publisher, logger: logger}
}

func (s *Service) PlaceOrder(ctx context.Context, req Request) (Response, error) {
	if req.UserID <= 0 || req.ProductID <= 0 || req.Quantity <= 0 {
		return Response{}, ErrInvalidRequest
	}

	s.logger.InfoContext(ctx, "checkout_started", "user_id", req.UserID, "product_id", req.ProductID, "qty", req.Quantity)
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return Response{}, err
	}
	defer tx.Rollback()

	// Checkout does not trust cache. The final stock decision happens while the
	// inventory row is locked in MySQL, which is what prevents overselling.
	stock, err := inventory.LockStockForUpdate(ctx, tx, req.ProductID)
	if err != nil {
		return Response{}, err
	}
	s.logger.InfoContext(ctx, "stock_locked", "product_id", req.ProductID)

	if stock.AvailableQty < req.Quantity {
		return Response{}, ErrInsufficientInventory
	}

	newQty := stock.AvailableQty - req.Quantity
	newVersion := stock.Version + 1
	if _, err := tx.ExecContext(ctx, `
		UPDATE inventory
		SET available_qty = ?, version = ?
		WHERE product_id = ?
	`, newQty, newVersion, req.ProductID); err != nil {
		return Response{}, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO orders (user_id, product_id, quantity, status)
		VALUES (?, ?, ?, 'confirmed')
	`, req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		return Response{}, err
	}
	orderID, err := result.LastInsertId()
	if err != nil {
		return Response{}, err
	}
	if err := tx.Commit(); err != nil {
		return Response{}, err
	}

	s.logger.InfoContext(ctx, "order_created", "order_id", orderID)
	if s.publisher != nil {
		// Publishing after commit avoids telling consumers about a write that
		// might still roll back.
		event := events.InventoryUpdatedEvent{
			EventID:   fmt.Sprintf("inventory-%d-%d", req.ProductID, time.Now().UnixNano()),
			ProductID: req.ProductID,
			NewQty:    newQty,
			Version:   newVersion,
			Reason:    "checkout",
			CreatedAt: time.Now().UTC(),
		}
		if err := s.publisher.PublishInventoryUpdated(ctx, event); err != nil {
			return Response{}, err
		}
		s.logger.InfoContext(ctx, "inventory_event_published", "product_id", req.ProductID, "version", newVersion)
	}

	return Response{
		OrderID:   orderID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Status:    "confirmed",
	}, nil
}

var (
	ErrInvalidRequest        = errors.New("invalid_checkout_request")
	ErrInsufficientInventory = errors.New("insufficient_inventory")
)
