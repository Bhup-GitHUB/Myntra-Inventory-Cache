package product

type Product struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Brand      string `json:"brand"`
	PriceCents int64  `json:"price"`
}

type Response struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Brand      string `json:"brand"`
	PriceCents int64  `json:"price"`
	Cache      string `json:"cache"`
}
