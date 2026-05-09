package product

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

func (r *Repository) GetByID(ctx context.Context, productID int64) (Product, error) {
	var product Product
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, brand, price_cents
		FROM products
		WHERE id = ?
	`, productID).Scan(&product.ID, &product.Name, &product.Brand, &product.PriceCents)
	return product, err
}
