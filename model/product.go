package model

type ProductListItem struct {
	ID             uint64  `db:"id" json:"id"`
	Name           string  `db:"name" json:"name"`
	ShopName       string  `db:"shop_name" json:"shop_name"`
	AvailableStock int64   `db:"available_stock" json:"available_stock"`
	Price          float64 `db:"price" json:"price"`
}

type ProductDetail struct {
	ID             uint64  `db:"id" json:"id"`
	Name           string  `db:"name" json:"name"`
	Description    string  `db:"description" json:"description,omitempty"`
	ShopID         uint64  `db:"shop_id" json:"shop_id"`
	ShopName       string  `db:"shop_name" json:"shop_name"`
	AvailableStock int64   `db:"available_stock" json:"available_stock"`
	Price          float64 `db:"price" json:"price"`
}

type ProductListResponse struct {
	Items      []ProductListItem `json:"items"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
}
