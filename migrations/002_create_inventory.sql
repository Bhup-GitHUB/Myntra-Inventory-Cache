-- +goose Up
CREATE TABLE IF NOT EXISTS inventory (
  product_id BIGINT PRIMARY KEY,
  available_qty INT NOT NULL,
  reserved_qty INT NOT NULL DEFAULT 0,
  version BIGINT NOT NULL DEFAULT 1,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_inventory_product FOREIGN KEY (product_id) REFERENCES products(id)
);

-- +goose Down
DROP TABLE IF EXISTS inventory;
