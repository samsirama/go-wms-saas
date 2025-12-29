package services

import (
	"context"

	"github.com/samsirama/go-wms-saas/internal/core/domain"
	"github.com/samsirama/go-wms-saas/internal/core/ports"
)

type ProductService struct {
	repo ports.ProductRepository
}

func NewProductService(repo ports.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, product *domain.Product) error {
	return s.repo.Create(ctx, product)
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	return s.repo.GetBySKU(ctx, sku)
}

func (s *ProductService) List(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *ProductService) Update(ctx context.Context, product *domain.Product) error {
	return s.repo.Update(ctx, product)
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
