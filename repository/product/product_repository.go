package product

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/muhammadheryan/e-commerce/model"
)

type SQL struct {
	conn *sqlx.DB
}

type ProductRepository interface {
	List(ctx context.Context, page, perPage int) ([]model.ProductListItem, int64, error)
	GetByID(ctx context.Context, id uint64) (*model.ProductDetail, error)
}

func NewProductRepository(conn *sqlx.DB) ProductRepository {
	return &SQL{conn: conn}
}

const (
	listProductsBase = `SELECT p.id, p.name, p.price, s.name as shop_name, COALESCE(SUM(ws.stock - ws.reserved),0) as available_stock
FROM product p
JOIN shop s ON p.shop_id = s.id
LEFT JOIN warehouse_stock ws ON ws.product_id = p.id
GROUP BY p.id, p.name, p.price, s.name`

	countProductsQuery = `SELECT COUNT(*) FROM product`

	getProductDetail = `SELECT p.id, p.name, p.description, p.price, s.id as shop_id, s.name as shop_name, COALESCE(SUM(ws.stock - ws.reserved),0) as available_stock
FROM product p
JOIN shop s ON p.shop_id = s.id
LEFT JOIN warehouse_stock ws ON ws.product_id = p.id
WHERE p.id = ?
GROUP BY p.id, p.name, p.description, p.price, s.id, s.name`
)

func (s *SQL) List(ctx context.Context, page, perPage int) ([]model.ProductListItem, int64, error) {
	offset := (page - 1) * perPage

	query := listProductsBase + " ORDER BY p.id LIMIT ? OFFSET ?"
	rows, err := s.conn.QueryxContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.ProductListItem, 0)
	for rows.Next() {
		var it model.ProductListItem
		if err := rows.StructScan(&it); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}

	// get total count
	var total int64
	if err := s.conn.GetContext(ctx, &total, countProductsQuery); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *SQL) GetByID(ctx context.Context, id uint64) (*model.ProductDetail, error) {
	var detail model.ProductDetail
	if err := s.conn.QueryRowxContext(ctx, getProductDetail, id).StructScan(&detail); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &detail, nil
}
