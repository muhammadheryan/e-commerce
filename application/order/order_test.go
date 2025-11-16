package order_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	apporder "github.com/muhammadheryan/e-commerce/application/order"
	"github.com/muhammadheryan/e-commerce/cmd/config"
	"github.com/muhammadheryan/e-commerce/constant"
	ordermocks "github.com/muhammadheryan/e-commerce/mocks/repository/order"
	txmocks "github.com/muhammadheryan/e-commerce/mocks/repository/tx"
	warehousemocks "github.com/muhammadheryan/e-commerce/mocks/repository/warehouse"
	"github.com/muhammadheryan/e-commerce/model"
	cerr "github.com/muhammadheryan/e-commerce/utils/errors"
	"github.com/stretchr/testify/mock"
)

// Note: order.go now checks if publisher is nil before calling PublishOrderExpiration
// So we can use nil publisher in tests without panicking

func TestOrderApp_CreateOrder(t *testing.T) {
	type fields struct {
		config        *config.Config
		txRepo        *txmocks.TxRepository
		orderRepo     *ordermocks.OrderRepository
		warehouseRepo *warehousemocks.WarehouseRepository
	}
	type args struct {
		ctx    context.Context
		userID uint64
		req    *model.OrderRequest
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockCall func(f fields)
		want     *model.OrderResponse
		wantErr  bool
		errCode  constant.ErrorType
	}{
		{
			name: "success: create order with single item",
			fields: fields{
				config: &config.Config{
					Order: config.OrderConfig{
						OrderExpiration: 30 * time.Minute,
					},
				},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: &model.OrderRequest{
					Items: []model.OrderItemRequest{
						{
							ProductID: 1,
							Quantity:  5,
						},
					},
				},
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("CommitTx", tx).Return(nil).Once()

				f.warehouseRepo.On("GetTotalAvailableStockTx", mock.Anything, tx, uint64(1)).Return(int64(100), nil).Once()

				f.orderRepo.On("InsertOrderTx", mock.Anything, tx, mock.MatchedBy(func(req *model.InsertOrderTxItem) bool {
					return req.UserID == 1 && req.Status == constant.OrderStatusPending
				})).Return(uint64(1), nil).Once()

				f.orderRepo.On("InsertOrderItemsTx", mock.Anything, tx, uint64(1), []model.OrderItemRequest{
					{ProductID: 1, Quantity: 5},
				}).Return(nil).Once()

				f.warehouseRepo.On("ReserveStockTx", mock.Anything, tx, mock.MatchedBy(func(req *model.ReserveRequest) bool {
					return req.OrderID == 1 && req.ProductID == 1 && req.Quantity == 5
				})).Return(nil).Once()
			},
			want: &model.OrderResponse{
				OrderID: 1,
			},
			wantErr: false,
		},
		{
			name: "error: empty items",
			fields: fields{
				config: &config.Config{
					Order: config.OrderConfig{
						OrderExpiration: 30 * time.Minute,
					},
				},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: &model.OrderRequest{
					Items: []model.OrderItemRequest{},
				},
			},
			mockCall: nil,
			want:     nil,
			wantErr:  true,
			errCode:  constant.ErrInvalidRequest,
		},
		{
			name: "error: insufficient stock",
			fields: fields{
				config: &config.Config{
					Order: config.OrderConfig{
						OrderExpiration: 30 * time.Minute,
					},
				},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: &model.OrderRequest{
					Items: []model.OrderItemRequest{
						{
							ProductID: 1,
							Quantity:  100,
						},
					},
				},
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.warehouseRepo.On("GetTotalAvailableStockTx", mock.Anything, tx, uint64(1)).Return(int64(50), nil).Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInsufficientStock,
		},
		{
			name: "error: BeginTx returns error",
			fields: fields{
				config: &config.Config{
					Order: config.OrderConfig{
						OrderExpiration: 30 * time.Minute,
					},
				},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: &model.OrderRequest{
					Items: []model.OrderItemRequest{
						{ProductID: 1, Quantity: 5},
					},
				},
			},
			mockCall: func(f fields) {
				f.txRepo.On("BeginTx", mock.Anything).Return(nil, errors.New("tx error")).Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInternal,
		},
		{
			name: "error: GetTotalAvailableStockTx returns error",
			fields: fields{
				config: &config.Config{
					Order: config.OrderConfig{
						OrderExpiration: 30 * time.Minute,
					},
				},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: &model.OrderRequest{
					Items: []model.OrderItemRequest{
						{ProductID: 1, Quantity: 5},
					},
				},
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.warehouseRepo.On("GetTotalAvailableStockTx", mock.Anything, tx, uint64(1)).Return(int64(0), errors.New("db error")).Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInternal,
		},
		{
			name: "error: ReserveStockTx returns insufficient stock error",
			fields: fields{
				config: &config.Config{
					Order: config.OrderConfig{
						OrderExpiration: 30 * time.Minute,
					},
				},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: &model.OrderRequest{
					Items: []model.OrderItemRequest{
						{ProductID: 1, Quantity: 5},
					},
				},
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.warehouseRepo.On("GetTotalAvailableStockTx", mock.Anything, tx, uint64(1)).Return(int64(100), nil).Once()

				f.orderRepo.On("InsertOrderTx", mock.Anything, tx, mock.Anything).Return(uint64(1), nil).Once()

				f.orderRepo.On("InsertOrderItemsTx", mock.Anything, tx, uint64(1), mock.Anything).Return(nil).Once()

				insufficientStockErr := cerr.SetCustomError(constant.ErrInsufficientStock)
				f.warehouseRepo.On("ReserveStockTx", mock.Anything, tx, mock.Anything).Return(insufficientStockErr).Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInsufficientStock,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockCall != nil {
				ttFields := tt.fields
				tt.mockCall(ttFields)
			}
			// Use nil publisher since order.go now checks for nil before calling
			app := apporder.NewOrderApp(tt.fields.config, tt.fields.txRepo, tt.fields.orderRepo, tt.fields.warehouseRepo, nil)

			got, err := app.CreateOrder(tt.args.ctx, tt.args.userID, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ce cerr.CustomError
				if !errors.As(err, &ce) {
					t.Fatalf("error type = %T, want CustomError", err)
				}
				if ce.ErrorCode() != constant.ErrorTypeCode[tt.errCode] {
					t.Fatalf("error code = %s, want %s", ce.ErrorCode(), constant.ErrorTypeCode[tt.errCode])
				}
				return
			}

			if got.OrderID != tt.want.OrderID {
				t.Fatalf("CreateOrder() OrderID = %v, want %v", got.OrderID, tt.want.OrderID)
			}
			if got.ExpiresAt.IsZero() {
				t.Fatal("CreateOrder() ExpiresAt should not be zero")
			}
		})
	}
}

func TestOrderApp_PayOrder(t *testing.T) {
	type fields struct {
		config        *config.Config
		txRepo        *txmocks.TxRepository
		orderRepo     *ordermocks.OrderRepository
		warehouseRepo *warehousemocks.WarehouseRepository
	}
	type args struct {
		ctx     context.Context
		orderID uint64
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockCall func(f fields)
		wantErr  bool
		errCode  constant.ErrorType
	}{
		{
			name: "success: pay order",
			fields: fields{
				config:        &config.Config{},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("CommitTx", tx).Return(nil).Once()

				f.orderRepo.On("GetOrderDetailTx", mock.Anything, tx, uint64(1)).Return(&model.OrderDetail{
					ID:     1,
					UserID: 1,
					Status: constant.OrderStatusPending,
				}, nil).Once()

				f.warehouseRepo.On("CommitReservationsTx", mock.Anything, tx, uint64(1)).Return(nil).Once()

				f.orderRepo.On("UpdateOrderStatusTx", mock.Anything, tx, uint64(1), int(constant.OrderStatusCompleted)).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "error: order not found",
			fields: fields{
				config:        &config.Config{},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 999,
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.orderRepo.On("GetOrderDetailTx", mock.Anything, tx, uint64(999)).Return(nil, errors.New("not found")).Once()
			},
			wantErr: true,
			errCode: constant.ErrInternal,
		},
		{
			name: "error: invalid order status (not pending)",
			fields: fields{
				config:        &config.Config{},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.orderRepo.On("GetOrderDetailTx", mock.Anything, tx, uint64(1)).Return(&model.OrderDetail{
					ID:     1,
					UserID: 1,
					Status: constant.OrderStatusCompleted,
				}, nil).Once()
			},
			wantErr: true,
			errCode: constant.ErrInvalidOrderStatus,
		},
		{
			name: "error: CommitReservationsTx returns error",
			fields: fields{
				config:        &config.Config{},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.orderRepo.On("GetOrderDetailTx", mock.Anything, tx, uint64(1)).Return(&model.OrderDetail{
					ID:     1,
					UserID: 1,
					Status: constant.OrderStatusPending,
				}, nil).Once()

				f.warehouseRepo.On("CommitReservationsTx", mock.Anything, tx, uint64(1)).Return(errors.New("commit error")).Once()
			},
			wantErr: true,
			errCode: constant.ErrInternal,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockCall != nil {
				ttFields := tt.fields
				tt.mockCall(ttFields)
			}
			app := apporder.NewOrderApp(tt.fields.config, tt.fields.txRepo, tt.fields.orderRepo, tt.fields.warehouseRepo, nil)

			err := app.PayOrder(tt.args.ctx, tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PayOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ce cerr.CustomError
				if !errors.As(err, &ce) {
					t.Fatalf("error type = %T, want CustomError", err)
				}
				if ce.ErrorCode() != constant.ErrorTypeCode[tt.errCode] {
					t.Fatalf("error code = %s, want %s", ce.ErrorCode(), constant.ErrorTypeCode[tt.errCode])
				}
			}
		})
	}
}

func TestOrderApp_CancelOrder(t *testing.T) {
	type fields struct {
		config        *config.Config
		txRepo        *txmocks.TxRepository
		orderRepo     *ordermocks.OrderRepository
		warehouseRepo *warehousemocks.WarehouseRepository
	}
	type args struct {
		ctx     context.Context
		orderID uint64
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockCall func(f fields)
		wantErr  bool
		errCode  constant.ErrorType
	}{
		{
			name: "success: cancel order",
			fields: fields{
				config:        &config.Config{},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("CommitTx", tx).Return(nil).Once()

				f.orderRepo.On("GetOrderDetailTx", mock.Anything, tx, uint64(1)).Return(&model.OrderDetail{
					ID:     1,
					UserID: 1,
					Status: constant.OrderStatusPending,
				}, nil).Once()

				f.warehouseRepo.On("ReleaseReservationsTx", mock.Anything, tx, uint64(1)).Return(nil).Once()

				f.orderRepo.On("UpdateOrderStatusTx", mock.Anything, tx, uint64(1), int(constant.OrderStatusCanceled)).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "error: order not found",
			fields: fields{
				config:        &config.Config{},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 999,
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.orderRepo.On("GetOrderDetailTx", mock.Anything, tx, uint64(999)).Return(nil, errors.New("not found")).Once()
			},
			wantErr: true,
			errCode: constant.ErrInternal,
		},
		{
			name: "error: invalid order status (not pending)",
			fields: fields{
				config:        &config.Config{},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.orderRepo.On("GetOrderDetailTx", mock.Anything, tx, uint64(1)).Return(&model.OrderDetail{
					ID:     1,
					UserID: 1,
					Status: constant.OrderStatusCompleted,
				}, nil).Once()
			},
			wantErr: true,
			errCode: constant.ErrInvalidOrderStatus,
		},
		{
			name: "error: ReleaseReservationsTx returns error",
			fields: fields{
				config:        &config.Config{},
				txRepo:        txmocks.NewTxRepository(t),
				orderRepo:     ordermocks.NewOrderRepository(t),
				warehouseRepo: warehousemocks.NewWarehouseRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			mockCall: func(f fields) {
				tx := &sqlx.Tx{}
				f.txRepo.On("BeginTx", mock.Anything).Return(tx, nil).Once()
				f.txRepo.On("RollbackTx", tx).Return(nil).Once()

				f.orderRepo.On("GetOrderDetailTx", mock.Anything, tx, uint64(1)).Return(&model.OrderDetail{
					ID:     1,
					UserID: 1,
					Status: constant.OrderStatusPending,
				}, nil).Once()

				f.warehouseRepo.On("ReleaseReservationsTx", mock.Anything, tx, uint64(1)).Return(errors.New("release error")).Once()
			},
			wantErr: true,
			errCode: constant.ErrInternal,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockCall != nil {
				ttFields := tt.fields
				tt.mockCall(ttFields)
			}
			app := apporder.NewOrderApp(tt.fields.config, tt.fields.txRepo, tt.fields.orderRepo, tt.fields.warehouseRepo, nil)

			err := app.CancelOrder(tt.args.ctx, tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CancelOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ce cerr.CustomError
				if !errors.As(err, &ce) {
					t.Fatalf("error type = %T, want CustomError", err)
				}
				if ce.ErrorCode() != constant.ErrorTypeCode[tt.errCode] {
					t.Fatalf("error code = %s, want %s", ce.ErrorCode(), constant.ErrorTypeCode[tt.errCode])
				}
			}
		})
	}
}
