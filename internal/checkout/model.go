package checkout

type Request struct {
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type Response struct {
	OrderID   int64  `json:"order_id"`
	ProductID int64  `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
}
