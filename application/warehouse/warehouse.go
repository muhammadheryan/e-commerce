package warehouse

import (
	"context"
	"database/sql"

	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	txrepo "github.com/muhammadheryan/e-commerce/repository/tx"
	warehouserepo "github.com/muhammadheryan/e-commerce/repository/warehouse"
	"github.com/muhammadheryan/e-commerce/utils/errors"
	"github.com/muhammadheryan/e-commerce/utils/logger"
	"go.uber.org/zap"
)

type WarehouseApp interface {
	ActivateWarehouse(ctx context.Context, warehouseID uint64) error
	DeactivateWarehouse(ctx context.Context, warehouseID uint64) error
	TransferStock(ctx context.Context, req *model.TransferStockRequest) error
}

type warehouseAppImpl struct {
	txRepo        txrepo.TxRepository
	warehouseRepo warehouserepo.WarehouseRepository
}

func NewWarehouseApp(txRepo txrepo.TxRepository, warehouseRepo warehouserepo.WarehouseRepository) WarehouseApp {
	return &warehouseAppImpl{
		txRepo:        txRepo,
		warehouseRepo: warehouseRepo,
	}
}

func (s *warehouseAppImpl) ActivateWarehouse(ctx context.Context, warehouseID uint64) error {
	// Check if warehouse exists
	warehouse, err := s.warehouseRepo.GetWarehouseByID(ctx, warehouseID)
	if err != nil {
		logger.Error("[ActivateWarehouse] get warehouse failed", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	if warehouse == nil {
		return errors.SetCustomError(constant.ErrNotFound)
	}

	// Update status to active
	err = s.warehouseRepo.UpdateWarehouseStatus(ctx, warehouseID, constant.WarehouseStatusActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.SetCustomError(constant.ErrNotFound)
		}
		logger.Error("[ActivateWarehouse] update status failed", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}

	return nil
}

func (s *warehouseAppImpl) DeactivateWarehouse(ctx context.Context, warehouseID uint64) error {
	// Check if warehouse exists
	warehouse, err := s.warehouseRepo.GetWarehouseByID(ctx, warehouseID)
	if err != nil {
		logger.Error("[DeactivateWarehouse] get warehouse failed", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	if warehouse == nil {
		return errors.SetCustomError(constant.ErrNotFound)
	}

	// Check if theres any reserved stock
	reservedStock, err := s.warehouseRepo.CheckReservedStock(ctx, warehouseID)
	if err != nil {
		logger.Error("[DeactivateWarehouse] check reserved stock failed", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	if reservedStock > 0 {
		return errors.SetCustomError(constant.ErrWarehouseHasReservedStock)
	}

	// Update status to inactive
	err = s.warehouseRepo.UpdateWarehouseStatus(ctx, warehouseID, constant.WarehouseStatusInactive)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.SetCustomError(constant.ErrNotFound)
		}
		logger.Error("[DeactivateWarehouse] update status failed", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}

	return nil
}

func (s *warehouseAppImpl) TransferStock(ctx context.Context, req *model.TransferStockRequest) error {
	// Validate request
	if req.FromWarehouseID == req.ToWarehouseID {
		return errors.SetCustomError(constant.ErrInvalidRequest)
	}
	if req.Quantity <= 0 {
		return errors.SetCustomError(constant.ErrInvalidRequest)
	}

	// Start transaction
	tx, err := s.txRepo.BeginTx(ctx)
	if err != nil {
		logger.Error("[TransferStock] begin tx failed", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	committed := false
	defer func() {
		if !committed {
			_ = s.txRepo.RollbackTx(tx)
		}
	}()

	// Transfer stock
	err = s.warehouseRepo.TransferStockTx(ctx, tx, req)
	if err != nil {
		logger.Error("[TransferStock] transfer stock failed", zap.String("error", err.Error()))
		if err.Error() == errors.SetCustomError(constant.ErrNotFound).Error() {
			return errors.SetCustomError(constant.ErrNotFound)
		}
		if err.Error() == errors.SetCustomError(constant.ErrInsufficientStock).Error() {
			return errors.SetCustomError(constant.ErrInsufficientStock)
		}
		return errors.SetCustomError(constant.ErrInternal)
	}

	// Commit transaction
	if err := s.txRepo.CommitTx(tx); err != nil {
		logger.Error("[TransferStock] commit tx failed", zap.String("error", err.Error()))
		return errors.SetCustomError(constant.ErrInternal)
	}
	committed = true

	return nil
}
