package services

import (
	"context"
	"fmt"

	"github.com/samsirama/go-wms-saas/internal/core/domain"
	"github.com/samsirama/go-wms-saas/internal/core/ports"
)

type ProductService struct {
	productRepo ports.ProductRepository
	stockRepo   ports.StockRepository
	txManager   ports.TransactionManager
}

func NewProductService(
	productRepo ports.ProductRepository,
	stockRepo ports.StockRepository,
	txManager ports.TransactionManager,
) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		stockRepo:   stockRepo,
		txManager:   txManager,
	}
}

func (s *ProductService) Create(ctx context.Context, product *domain.Product) error {
	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.productRepo.Create(txCtx, product); err != nil {
			return fmt.Errorf("create product: %w", err)
		}

		stock := &domain.Stock{
			ProductID: product.ID,
			Quantity:  0,
			Version:   1,
		}

		if err := s.stockRepo.Create(txCtx, stock); err != nil {
			return fmt.Errorf("create stock: %w", err)
		}

		return nil
	})
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	return s.productRepo.GetBySKU(ctx, sku)
}

func (s *ProductService) List(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	return s.productRepo.List(ctx, limit, offset)
}

func (s *ProductService) Update(ctx context.Context, product *domain.Product) error {
	return s.productRepo.Update(ctx, product)
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.productRepo.Delete(ctx, id)
}
