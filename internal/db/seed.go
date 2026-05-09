package db

import (
	"context"
	"database/sql"
)

func ResetDemoData(ctx context.Context, conn *sql.DB) error {
	statements := []string{
		"DELETE FROM orders",
		"DELETE FROM inventory",
		"DELETE FROM products",
		"ALTER TABLE orders AUTO_INCREMENT = 9001",
		"INSERT INTO products (id, name, brand, price_cents) VALUES (101, 'Running Shoes', 'DemoBrand', 2999), (102, 'Training T-Shirt', 'DemoBrand', 999), (103, 'City Backpack', 'DemoGear', 1999)",
		"INSERT INTO inventory (product_id, available_qty, reserved_qty, version) VALUES (101, 25, 3, 1), (102, 80, 0, 1), (103, 15, 1, 1)",
	}
	for _, statement := range statements {
		if _, err := conn.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
