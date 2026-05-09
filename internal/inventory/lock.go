package inventory

import (
	"context"
	"database/sql"
)

type LockedStock struct {
	AvailableQty int
	Version      int64
}

func LockStockForUpdate(ctx context.Context, tx *sql.Tx, productID int64) (LockedStock, error) {
	var stock LockedStock
	err := tx.QueryRowContext(ctx, `
		SELECT available_qty, version
		FROM inventory
		WHERE product_id = ?
		FOR UPDATE
	`, productID).Scan(&stock.AvailableQty, &stock.Version)
	return stock, err
}
