package warehouse

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	"github.com/muhammadheryan/e-commerce/utils/errors"
)

type WarehouseRepository interface {
	GetTotalAvailableStockTx(ctx context.Context, tx *sqlx.Tx, productID uint64) (int64, error)
	ReserveStockTx(ctx context.Context, tx *sqlx.Tx, req *model.ReserveRequest) error
	GetReservationsByOrderTx(ctx context.Context, tx *sqlx.Tx, orderID uint64) ([]model.Reservation, error)
	CommitReservationsTx(ctx context.Context, tx *sqlx.Tx, orderID uint64) error
	ReleaseReservationsTx(ctx context.Context, tx *sqlx.Tx, orderID uint64) error
}

type SQL struct {
	conn *sqlx.DB
}

func NewWarehouseRepository(conn *sqlx.DB) WarehouseRepository {
	return &SQL{conn: conn}
}

func (r *SQL) GetTotalAvailableStockTx(ctx context.Context, tx *sqlx.Tx, productID uint64) (int64, error) {
	var total sql.NullInt64
	q := "SELECT COALESCE(SUM(ws.stock - ws.reserved),0) as total FROM warehouse_stock ws JOIN warehouse w ON ws.warehouse_id = w.id WHERE ws.product_id = ? AND w.status = ?"
	if err := tx.GetContext(ctx, &total, q, productID, constant.WarehouseStatusActive); err != nil {
		return 0, err
	}
	if !total.Valid {
		return 0, nil
	}
	return total.Int64, nil
}

func (r *SQL) ReserveStockTx(ctx context.Context, tx *sqlx.Tx, req *model.ReserveRequest) error {
	// Lock rows for this product to avoid races
	rows, err := tx.QueryxContext(ctx, "SELECT ws.id, ws.warehouse_id, ws.stock, ws.reserved FROM warehouse_stock ws JOIN warehouse w ON ws.warehouse_id = w.id WHERE ws.product_id = ? AND w.status = ? FOR UPDATE", req.ProductID, constant.WarehouseStatusActive)
	if err != nil {
		return err
	}
	defer rows.Close()

	type ws struct {
		ID          int64 `db:"id"`
		WarehouseID int64 `db:"warehouse_id"`
		Stock       int64 `db:"stock"`
		Reserved    int64 `db:"reserved"`
	}

	needed := int64(req.Quantity)
	for rows.Next() {
		var w ws
		if err := rows.StructScan(&w); err != nil {
			return err
		}
		avail := w.Stock - w.Reserved
		if avail <= 0 {
			continue
		}
		alloc := avail
		if alloc > needed {
			alloc = needed
		}
		// update reserved
		if _, err := tx.ExecContext(ctx, "UPDATE warehouse_stock SET reserved = reserved + ? WHERE id = ?", alloc, w.ID); err != nil {
			return err
		}
		// insert reservation record with expires_at
		if _, err := tx.ExecContext(ctx, "INSERT INTO stock_reservation (order_id, warehouse_id, product_id, quantity, expires_at) VALUES (?, ?, ?, ?, ?)", req.OrderID, w.WarehouseID, req.ProductID, alloc, req.ExpiresAt); err != nil {
			return err
		}
		needed -= alloc
		if needed <= 0 {
			break
		}
	}

	if needed > 0 {
		return errors.SetCustomError(constant.ErrInsufficientStock)
	}

	return nil
}

func (r *SQL) GetReservationsByOrderTx(ctx context.Context, tx *sqlx.Tx, orderID uint64) ([]model.Reservation, error) {
	rows, err := tx.QueryxContext(ctx, "SELECT id, warehouse_id, product_id, quantity FROM stock_reservation WHERE order_id = ? FOR UPDATE", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]model.Reservation, 0)
	for rows.Next() {
		var rr model.Reservation
		if err := rows.StructScan(&rr); err != nil {
			return nil, err
		}
		res = append(res, rr)
	}
	return res, nil
}

func (r *SQL) CommitReservationsTx(ctx context.Context, tx *sqlx.Tx, orderID uint64) error {
	reservations, err := r.GetReservationsByOrderTx(ctx, tx, orderID)
	if err != nil {
		return err
	}
	for _, reservation := range reservations {
		// decrease stock and reserved
		if _, err := tx.ExecContext(ctx, "UPDATE warehouse_stock SET stock = stock - ?, reserved = reserved - ? WHERE warehouse_id = ? AND product_id = ?", reservation.Quantity, reservation.Quantity, reservation.WarehouseID, reservation.ProductID); err != nil {
			return err
		}
		// delete reservation row
		if _, err := tx.ExecContext(ctx, "DELETE FROM stock_reservation WHERE id = ?", reservation.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQL) ReleaseReservationsTx(ctx context.Context, tx *sqlx.Tx, orderID uint64) error {
	reservations, err := r.GetReservationsByOrderTx(ctx, tx, orderID)
	if err != nil {
		return err
	}
	for _, rr := range reservations {
		// decrease reserved only
		if _, err := tx.ExecContext(ctx, "UPDATE warehouse_stock SET reserved = reserved - ? WHERE warehouse_id = ? AND product_id = ?", rr.Quantity, rr.WarehouseID, rr.ProductID); err != nil {
			return err
		}
		// delete reservation row
		if _, err := tx.ExecContext(ctx, "DELETE FROM stock_reservation WHERE id = ?", rr.ID); err != nil {
			return err
		}
	}
	return nil
}
