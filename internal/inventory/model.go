package inventory

type Inventory struct {
	ProductID    int64 `json:"product_id"`
	AvailableQty int   `json:"available_qty"`
	ReservedQty  int   `json:"reserved_qty"`
	Version      int64 `json:"version"`
}

type Response struct {
	ProductID    int64    `json:"product_id"`
	AvailableQty int      `json:"available_qty"`
	ReservedQty  int      `json:"reserved_qty"`
	CacheSource  string   `json:"cache_source"`
	RequestPath  []string `json:"request_path"`
}

type InvalidateResponse struct {
	ProductID   int64 `json:"product_id"`
	Invalidated bool  `json:"invalidated"`
}
