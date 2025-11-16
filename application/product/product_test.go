package product_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	appproduct "github.com/muhammadheryan/e-commerce/application/product"
	"github.com/muhammadheryan/e-commerce/constant"
	productmocks "github.com/muhammadheryan/e-commerce/mocks/repository/product"
	"github.com/muhammadheryan/e-commerce/model"
	cerr "github.com/muhammadheryan/e-commerce/utils/errors"
	"github.com/stretchr/testify/mock"
)

func TestProductApp_ListProducts(t *testing.T) {
	type fields struct {
		productRepo *productmocks.ProductRepository
	}
	type args struct {
		ctx     context.Context
		page    int
		perPage int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockCall func(f fields)
		want     *model.ProductListResponse
		wantErr  bool
	}{
		{
			name: "success: list products with pagination",
			fields: fields{
				productRepo: productmocks.NewProductRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				page:    1,
				perPage: 10,
			},
			mockCall: func(f fields) {
				items := []model.ProductListItem{
					{
						ID:             1,
						Name:           "Product 1",
						ShopName:       "Shop A",
						AvailableStock: 100,
						Price:          50000.0,
					},
					{
						ID:             2,
						Name:           "Product 2",
						ShopName:       "Shop B",
						AvailableStock: 50,
						Price:          75000.0,
					},
				}
				f.productRepo.
					On("List", mock.Anything, 1, 10).
					Return(items, int64(2), nil).
					Once()
			},
			want: &model.ProductListResponse{
				Items: []model.ProductListItem{
					{
						ID:             1,
						Name:           "Product 1",
						ShopName:       "Shop A",
						AvailableStock: 100,
						Price:          50000.0,
					},
					{
						ID:             2,
						Name:           "Product 2",
						ShopName:       "Shop B",
						AvailableStock: 50,
						Price:          75000.0,
					},
				},
				TotalCount: 2,
				Page:       1,
				PerPage:    10,
			},
			wantErr: false,
		},
		{
			name: "success: default page and perPage when zero or negative",
			fields: fields{
				productRepo: productmocks.NewProductRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				page:    0,
				perPage: 0,
			},
			mockCall: func(f fields) {
				f.productRepo.
					On("List", mock.Anything, 1, 10).
					Return([]model.ProductListItem{}, int64(0), nil).
					Once()
			},
			want: &model.ProductListResponse{
				Items:      []model.ProductListItem{},
				TotalCount: 0,
				Page:       1,
				PerPage:    10,
			},
			wantErr: false,
		},
		{
			name: "success: negative page defaults to 1",
			fields: fields{
				productRepo: productmocks.NewProductRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				page:    -1,
				perPage: 5,
			},
			mockCall: func(f fields) {
				f.productRepo.
					On("List", mock.Anything, 1, 5).
					Return([]model.ProductListItem{}, int64(0), nil).
					Once()
			},
			want: &model.ProductListResponse{
				Items:      []model.ProductListItem{},
				TotalCount: 0,
				Page:       1,
				PerPage:    5,
			},
			wantErr: false,
		},
		{
			name: "error: repository List returns error",
			fields: fields{
				productRepo: productmocks.NewProductRepository(t),
			},
			args: args{
				ctx:     context.Background(),
				page:    1,
				perPage: 10,
			},
			mockCall: func(f fields) {
				f.productRepo.
					On("List", mock.Anything, 1, 10).
					Return(nil, int64(0), errors.New("db error")).
					Once()
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockCall != nil {
				ttFields := tt.fields
				tt.mockCall(ttFields)
			}
			app := appproduct.NewProductApp(tt.fields.productRepo)

			got, err := app.ListProducts(tt.args.ctx, tt.args.page, tt.args.perPage)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListProducts() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ce cerr.CustomError
				if !errors.As(err, &ce) {
					t.Fatalf("error type = %T, want CustomError", err)
				}
				if ce.ErrorCode() != constant.ErrorTypeCode[constant.ErrInternal] {
					t.Fatalf("error code = %s, want %s", ce.ErrorCode(), constant.ErrorTypeCode[constant.ErrInternal])
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ListProducts() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestProductApp_GetProduct(t *testing.T) {
	type fields struct {
		productRepo *productmocks.ProductRepository
	}
	type args struct {
		ctx context.Context
		id  uint64
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockCall func(f fields)
		want     *model.ProductDetail
		wantErr  bool
	}{
		{
			name: "success: get product by id",
			fields: fields{
				productRepo: productmocks.NewProductRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			mockCall: func(f fields) {
				f.productRepo.
					On("GetByID", mock.Anything, uint64(1)).
					Return(&model.ProductDetail{
						ID:             1,
						Name:           "Product 1",
						Description:    "Product description",
						ShopID:         10,
						ShopName:       "Shop A",
						AvailableStock: 100,
						Price:          50000.0,
					}, nil).
					Once()
			},
			want: &model.ProductDetail{
				ID:             1,
				Name:           "Product 1",
				Description:    "Product description",
				ShopID:         10,
				ShopName:       "Shop A",
				AvailableStock: 100,
				Price:          50000.0,
			},
			wantErr: false,
		},
		{
			name: "error: repository GetByID returns error",
			fields: fields{
				productRepo: productmocks.NewProductRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  999,
			},
			mockCall: func(f fields) {
				f.productRepo.
					On("GetByID", mock.Anything, uint64(999)).
					Return(nil, errors.New("db error")).
					Once()
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockCall != nil {
				ttFields := tt.fields
				tt.mockCall(ttFields)
			}
			app := appproduct.NewProductApp(tt.fields.productRepo)

			got, err := app.GetProduct(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetProduct() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ce cerr.CustomError
				if !errors.As(err, &ce) {
					t.Fatalf("error type = %T, want CustomError", err)
				}
				if ce.ErrorCode() != constant.ErrorTypeCode[constant.ErrInternal] {
					t.Fatalf("error code = %s, want %s", ce.ErrorCode(), constant.ErrorTypeCode[constant.ErrInternal])
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetProduct() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
