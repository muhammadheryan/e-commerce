package order

import (
	"context"
	"time"

	"github.com/muhammadheryan/e-commerce/cmd/config"
	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	orderrepo "github.com/muhammadheryan/e-commerce/repository/order"
	txrepo "github.com/muhammadheryan/e-commerce/repository/tx"
	warehouserepo "github.com/muhammadheryan/e-commerce/repository/warehouse"
	"github.com/muhammadheryan/e-commerce/thirdparty/rabbitmq"
	"github.com/muhammadheryan/e-commerce/utils/errors"
	"github.com/muhammadheryan/e-commerce/utils/logger"
	"go.uber.org/zap"
)

type OrderApp interface {
	CreateOrder(ctx context.Context, UserID uint64, req *model.OrderRequest) (*model.OrderResponse, error)
	PayOrder(ctx context.Context, orderID uint64) error
	CancelOrder(ctx context.Context, orderID uint64) error
}

type orderAppImpl struct {
	config        *config.Config
	txRepo        txrepo.TxRepository
	orderRepo     orderrepo.OrderRepository
	warehouseRepo warehouserepo.WarehouseRepository
	publisher     *rabbitmq.Publisher
}

func NewOrderApp(config *config.Config, txRepo txrepo.TxRepository, orderRepo orderrepo.OrderRepository, warehouseRepo warehouserepo.WarehouseRepository, publisher *rabbitmq.Publisher) OrderApp {
	return &orderAppImpl{config: config, txRepo: txRepo, orderRepo: orderRepo, warehouseRepo: warehouseRepo, publisher: publisher}
}

func (s *orderAppImpl) CreateOrder(ctx context.Context, UserID uint64, req *model.OrderRequest) (*model.OrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, errors.SetCustomError(constant.ErrInvalidRequest)
	}

	tx, err := s.txRepo.BeginTx(ctx)
	if err != nil {
		logger.Error("[CreateOrder] begin tx", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}
	committed := false
	defer func() {
		if !committed {
			_ = s.txRepo.RollbackTx(tx)
		}
	}()

	// validate stock for each item
	for _, item := range req.Items {
		total, err := s.warehouseRepo.GetTotalAvailableStockTx(ctx, tx, item.ProductID)
		if err != nil {
			logger.Error("[CreateOrder] get total stock", zap.String("error", err.Error()))
			return nil, errors.SetCustomError(constant.ErrInternal)
		}
		if total < int64(item.Quantity) {
			logger.Info("[CreateOrder] insufficient stock", zap.Uint64("product_id", item.ProductID), zap.Int("need", item.Quantity), zap.Int64("available", total))
			return nil, errors.SetCustomError(constant.ErrInsufficientStock)
		}
	}

	// insert order
	expiresAt := time.Now().Add(s.config.Order.OrderExpiration)
	orderID, err := s.orderRepo.InsertOrderTx(ctx, tx, &model.InsertOrderTxItem{
		UserID:    UserID,
		Status:    constant.OrderStatusPending,
		ExpiresAT: expiresAt,
	})
	if err != nil {
		logger.Error("[CreateOrder] insert order", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	// insert items
	if err := s.orderRepo.InsertOrderItemsTx(ctx, tx, orderID, req.Items); err != nil {
		logger.Error("[CreateOrder] insert items", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	// reserve stock per item
	for _, item := range req.Items {
		req := &model.ReserveRequest{
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			ExpiresAt: expiresAt,
		}
		if err := s.warehouseRepo.ReserveStockTx(ctx, tx, req); err != nil {
			if err.Error() == errors.SetCustomError(constant.ErrInsufficientStock).Error() {
				return nil, errors.SetCustomError(constant.ErrInsufficientStock)
			}
			logger.Error("[CreateOrder] reserve stock", zap.String("error", err.Error()))
			return nil, errors.SetCustomError(constant.ErrInternal)
		}
	}

	if err := s.txRepo.CommitTx(tx); err != nil {
		logger.Error("[CreateOrder] commit tx", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}
	committed = true
	// Publish order expiration message to RabbitMQ
	msg := rabbitmq.OrderExpirationMessage{
		OrderID:   orderID,
		UserID:    UserID,
		ExpiresAt: expiresAt,
	}
	if err := s.publisher.PublishOrderExpiration(msg); err != nil {
		logger.Error("[CreateOrder] publish order expiration", zap.String("error", err.Error()))
	}

	return &model.OrderResponse{
		OrderID:   orderID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *orderAppImpl) PayOrder(ctx context.Context, orderID uint64) error {
	tx, err := s.txRepo.BeginTx(ctx)
	if err != nil {
		logger.Error("[PayOrder] begin tx", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	committed := false
	defer func() {
		if !committed {
			_ = s.txRepo.RollbackTx(tx)
		}
	}()

	// get order detail and validate status and ownership
	orderDetail, err := s.orderRepo.GetOrderDetailTx(ctx, tx, orderID)
	if err != nil {
		logger.Error("[PayOrder] get order detail", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}

	// verify status is pending
	if orderDetail.Status != constant.OrderStatusPending {
		return errors.SetCustomError(constant.ErrInvalidOrderStatus)
	}

	// commit reservations to decrease stock and reserved
	if err := s.warehouseRepo.CommitReservationsTx(ctx, tx, orderID); err != nil {
		logger.Error("[PayOrder] commit reservations", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}

	// update order status to completed
	if err := s.orderRepo.UpdateOrderStatusTx(ctx, tx, orderID, int(constant.OrderStatusCompleted)); err != nil {
		logger.Error("[PayOrder] update status", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}

	if err := s.txRepo.CommitTx(tx); err != nil {
		logger.Error("[PayOrder] commit tx", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	committed = true
	return nil
}

func (s *orderAppImpl) CancelOrder(ctx context.Context, orderID uint64) error {
	tx, err := s.txRepo.BeginTx(ctx)
	if err != nil {
		logger.Error("[CancelOrder] begin tx", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	committed := false
	defer func() {
		if !committed {
			_ = s.txRepo.RollbackTx(tx)
		}
	}()

	// get order detail and validate status and ownership
	orderDetail, err := s.orderRepo.GetOrderDetailTx(ctx, tx, orderID)
	if err != nil {
		logger.Error("[CancelOrder] get order detail", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}

	// verify status is pending
	if orderDetail.Status != constant.OrderStatusPending {
		return errors.SetCustomError(constant.ErrInvalidOrderStatus)
	}

	// release reservations to decrease reserved only
	if err := s.warehouseRepo.ReleaseReservationsTx(ctx, tx, orderID); err != nil {
		logger.Error("[CancelOrder] release reservations", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}

	// update order status to canceled
	if err := s.orderRepo.UpdateOrderStatusTx(ctx, tx, orderID, int(constant.OrderStatusCanceled)); err != nil {
		logger.Error("[CancelOrder] update status", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}

	if err := s.txRepo.CommitTx(tx); err != nil {
		logger.Error("[CancelOrder] commit tx", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	committed = true
	return nil
}
