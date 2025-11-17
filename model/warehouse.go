package model

import (
	"time"

	"github.com/muhammadheryan/e-commerce/constant"
)

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

type WarehouseEntity struct {
	ID        uint64                   `db:"id" json:"id"`
	ShopID    uint64                   `db:"shop_id" json:"shop_id"`
	Name      string                   `db:"name" json:"name"`
	Status    constant.WarehouseStatus `db:"status" json:"status"`
	CreatedAt time.Time                `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time               `db:"updated_at" json:"updated_at,omitempty"`
}

type WarehouseStock struct {
	ID          uint64 `db:"id" json:"id"`
	WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id"`
	ProductID   uint64 `db:"product_id" json:"product_id"`
	Stock       int64  `db:"stock" json:"stock"`
	Reserved    int64  `db:"reserved" json:"reserved"`
}

type TransferStockRequest struct {
	ProductID       uint64
	FromWarehouseID uint64
	ToWarehouseID   uint64
	Quantity        int
}

type TransferStockHTTPRequest struct {
	ProductID       uint64 `json:"product_id" validate:"required"`
	FromWarehouseID uint64 `json:"from_warehouse_id" validate:"required"`
	ToWarehouseID   uint64 `json:"to_warehouse_id" validate:"required"`
	Quantity        int    `json:"quantity" validate:"required,gt=0"`
}
