package product

import (
	"context"

	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	productRepo "github.com/muhammadheryan/e-commerce/repository/product"
	"github.com/muhammadheryan/e-commerce/utils/errors"
	"github.com/muhammadheryan/e-commerce/utils/logger"
	"go.uber.org/zap"
)

type ProductApp interface {
	ListProducts(ctx context.Context, page, perPage int) (*model.ProductListResponse, error)
	GetProduct(ctx context.Context, id uint64) (*model.ProductDetail, error)
}

type productAppImpl struct {
	productRepo productRepo.ProductRepository
}

func NewProductApp(productRepo productRepo.ProductRepository) ProductApp {
	return &productAppImpl{productRepo: productRepo}
}

func (s *productAppImpl) ListProducts(ctx context.Context, page, perPage int) (*model.ProductListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}

	items, total, err := s.productRepo.List(ctx, page, perPage)
	if err != nil {
		logger.Error("[ListProducts] error productRepo.List", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	return &model.ProductListResponse{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PerPage:    perPage,
	}, nil
}

func (s *productAppImpl) GetProduct(ctx context.Context, id uint64) (*model.ProductDetail, error) {
	result, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("[GetProduct] error productRepo.GetByID", zap.String("error", err.Error()))
		return nil, errors.SetCustomError(constant.ErrInternal)
	}

	return result, nil
}
