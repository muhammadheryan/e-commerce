package model

import (
	"time"

	"github.com/muhammadheryan/e-commerce/constant"
)

type OrderItemRequest struct {
	ProductID uint64 `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}

type OrderRequest struct {
	UserID uint64
	Items  []OrderItemRequest `json:"items" validate:"required,dive,required"`
}

type OrderResponse struct {
	OrderID   uint64    `json:"order_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type InsertOrderTxItem struct {
	UserID    uint64
	Status    constant.OrderStatus
	ExpiresAT time.Time
}

type OrderDetail struct {
	ID     uint64               `db:"id"`
	UserID uint64               `db:"user_id"`
	Status constant.OrderStatus `db:"status"`
}
