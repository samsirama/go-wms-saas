package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/samsirama/go-wms-saas/internal/core/domain"
	"github.com/samsirama/go-wms-saas/internal/core/ports"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock available")
	ErrVersionMismatch   = errors.New("version mismatch: stock was modified by another transaction")
	ErrProductNotFound   = errors.New("product not found")
	ErrInvalidQuantity   = errors.New("quantity must be greater than zero")
)

type StockService struct {
	stockRepo ports.StockRepository
	txManager ports.TransactionManager
}

func NewStockService(stockRepo ports.StockRepository, txManager ports.TransactionManager) *StockService {
	return &StockService{
		stockRepo: stockRepo,
		txManager: txManager,
	}
}

// ReserveStock decreases inventory for an order.
// Uses optimistic locking to prevent race conditions.
func (s *StockService) ReserveStock(ctx context.Context, productID string, quantity int64, orderID string) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		stock, err := s.stockRepo.GetByProductID(txCtx, productID)
		if err != nil {
			return fmt.Errorf("failed to get stock: %w", err)
		}

		if stock.Quantity < quantity {
			return ErrInsufficientStock
		}

		previousQty := stock.Quantity

		newVersion, err := s.stockRepo.UpdateWithOptimisticLock(txCtx, stock, -quantity)
		if err != nil {
			return fmt.Errorf("failed to update stock: %w", err)
		}

		mutation := &domain.StockMutation{
			ProductID:     productID,
			MutationType:  domain.MutationTypeReservation,
			Quantity:      quantity,
			PreviousQty:   previousQty,
			NewQty:        previousQty - quantity,
			ReferenceID:   orderID,
			ReferenceType: domain.ReferenceTypeOrder,
			Notes:         fmt.Sprintf("Reserved %d units for order %s (v%d)", quantity, orderID, newVersion),
		}

		if err := s.stockRepo.CreateMutation(txCtx, mutation); err != nil {
			return fmt.Errorf("failed to create mutation: %w", err)
		}

		return nil
	})
}

func (s *StockService) ReleaseStock(ctx context.Context, productID string, quantity int64, orderID string) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		stock, err := s.stockRepo.GetByProductID(txCtx, productID)
		if err != nil {
			return fmt.Errorf("failed to get stock: %w", err)
		}

		previousQty := stock.Quantity

		newVersion, err := s.stockRepo.UpdateWithOptimisticLock(txCtx, stock, quantity)
		if err != nil {
			return fmt.Errorf("failed to update stock: %w", err)
		}

		mutation := &domain.StockMutation{
			ProductID:     productID,
			MutationType:  domain.MutationTypeRelease,
			Quantity:      quantity,
			PreviousQty:   previousQty,
			NewQty:        previousQty + quantity,
			ReferenceID:   orderID,
			ReferenceType: domain.ReferenceTypeOrder,
			Notes:         fmt.Sprintf("Released %d units from order %s (v%d)", quantity, orderID, newVersion),
		}

		if err := s.stockRepo.CreateMutation(txCtx, mutation); err != nil {
			return fmt.Errorf("failed to create mutation: %w", err)
		}

		return nil
	})
}

func (s *StockService) GetStockLevel(ctx context.Context, productID string) (*domain.Stock, error) {
	return s.stockRepo.GetByProductID(ctx, productID)
}

func (s *StockService) GetMutationHistory(ctx context.Context, productID string, limit, offset int) ([]*domain.StockMutation, error) {
	return s.stockRepo.GetMutationHistory(ctx, productID, limit, offset)
}
