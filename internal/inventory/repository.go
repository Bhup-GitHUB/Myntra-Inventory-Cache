package inventory

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByProductID(ctx context.Context, productID int64) (Inventory, error) {
	var inventory Inventory
	err := r.db.QueryRowContext(ctx, `
		SELECT product_id, available_qty, reserved_qty, version
		FROM inventory
		WHERE product_id = ?
	`, productID).Scan(&inventory.ProductID, &inventory.AvailableQty, &inventory.ReservedQty, &inventory.Version)
	return inventory, err
}
