package order

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/muhammadheryan/e-commerce/model"
)

type SQL struct {
	conn *sqlx.DB
}

type OrderRepository interface {
	InsertOrderTx(ctx context.Context, tx *sqlx.Tx, req *model.InsertOrderTxItem) (uint64, error)
	InsertOrderItemsTx(ctx context.Context, tx *sqlx.Tx, orderID uint64, items []model.OrderItemRequest) error
	UpdateOrderStatusTx(ctx context.Context, tx *sqlx.Tx, orderID uint64, status int) error
	GetOrderDetailTx(ctx context.Context, tx *sqlx.Tx, orderID uint64) (*model.OrderDetail, error)
}

func NewOrderRepository(conn *sqlx.DB) OrderRepository {
	return &SQL{conn: conn}
}

func (r *SQL) InsertOrderTx(ctx context.Context, tx *sqlx.Tx, req *model.InsertOrderTxItem) (uint64, error) {
	res, err := tx.ExecContext(ctx, "INSERT INTO `order` (user_id, status, expires_at) VALUES (?, ?, ?)", req.UserID, req.Status, req.ExpiresAT)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *SQL) InsertOrderItemsTx(ctx context.Context, tx *sqlx.Tx, orderID uint64, items []model.OrderItemRequest) error {
	q := "INSERT INTO order_item (order_id, product_id, quantity) VALUES (?, ?, ?)"
	for _, it := range items {
		if _, err := tx.ExecContext(ctx, q, orderID, it.ProductID, it.Quantity); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQL) UpdateOrderStatusTx(ctx context.Context, tx *sqlx.Tx, orderID uint64, status int) error {
	_, err := tx.ExecContext(ctx, "UPDATE `order` SET status = ? WHERE id = ?", status, orderID)
	return err
}

func (r *SQL) GetOrderDetailTx(ctx context.Context, tx *sqlx.Tx, orderID uint64) (*model.OrderDetail, error) {
	var detail model.OrderDetail
	row := tx.QueryRowxContext(ctx, "SELECT id, user_id, status FROM `order` WHERE id = ?", orderID)
	if err := row.StructScan(&detail); err != nil {
		return nil, err
	}
	return &detail, nil
}
