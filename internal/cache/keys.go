package cache

import "fmt"

func InventoryKey(productID int64) string {
	return fmt.Sprintf("inventory:%d", productID)
}

func ProductKey(productID int64) string {
	return fmt.Sprintf("product:%d", productID)
}
