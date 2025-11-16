package model

import "time"

type ReserveRequest struct {
	OrderID   uint64
	ProductID uint64
	Quantity  int
	ExpiresAt time.Time
}

type Reservation struct {
	ID          int64  `db:"id"`
	WarehouseID int64  `db:"warehouse_id"`
	ProductID   uint64 `db:"product_id"`
	Quantity    int64  `db:"quantity"`
}
