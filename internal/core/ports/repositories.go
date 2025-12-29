package ports

import (
	"context"

	"github.com/samsirama/go-wms-saas/internal/core/domain"
)

type StockRepository interface {
	Create(ctx context.Context, stock *domain.Stock) error
	GetByProductID(ctx context.Context, productID string) (*domain.Stock, error)

	// UpdateWithOptimisticLock compares version before updating.
	// Returns new version on success, error on version mismatch.
	UpdateWithOptimisticLock(ctx context.Context, stock *domain.Stock, quantityDelta int64) (newVersion int64, err error)

	CreateMutation(ctx context.Context, mutation *domain.StockMutation) error
	GetMutationHistory(ctx context.Context, productID string, limit, offset int) ([]*domain.StockMutation, error)
}

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*domain.Product, error)
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
