package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/samsirama/go-wms-saas/internal/core/domain"
)

type StockRepo struct {
	db *PostgresDB
}

func NewStockRepo(pool *pgxpool.Pool) *StockRepo {
	return &StockRepo{
		db: NewPostgresDB(pool),
	}
}

func (r *StockRepo) Create(ctx context.Context, stock *domain.Stock) error {
	q := `
		INSERT INTO stocks (id, product_id, quantity, version, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	conn := r.db.getConn(ctx)
	row := conn.QueryRow(ctx, q, stock.ProductID, stock.Quantity, stock.Version)

	err := row.Scan(&stock.ID, &stock.CreatedAt, &stock.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert stock: %w", err)
	}

	return nil
}

func (r *StockRepo) GetByProductID(ctx context.Context, productID string) (*domain.Stock, error) {
	q := `
		SELECT id, product_id, quantity, version, created_at, updated_at
		FROM stocks
		WHERE product_id = $1
	`

	conn := r.db.getConn(ctx)
	row := conn.QueryRow(ctx, q, productID)

	var s domain.Stock
	err := row.Scan(&s.ID, &s.ProductID, &s.Quantity, &s.Version, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan stock: %w", err)
	}

	return &s, nil
}

func (r *StockRepo) UpdateWithOptimisticLock(ctx context.Context, stock *domain.Stock, quantityDelta int64) (int64, error) {
	q := `
		UPDATE stocks
		SET quantity = quantity + $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND version = $3
		RETURNING version
	`

	conn := r.db.getConn(ctx)
	row := conn.QueryRow(ctx, q, quantityDelta, stock.ProductID, stock.Version)

	var newVersion int64
	err := row.Scan(&newVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrOptimisticLock
		}
		return 0, fmt.Errorf("update stock: %w", err)
	}

	return newVersion, nil
}

func (r *StockRepo) CreateMutation(ctx context.Context, m *domain.StockMutation) error {
	q := `
		INSERT INTO stock_mutations (
			id, product_id, mutation_type, quantity, previous_qty, new_qty,
			reference_id, reference_type, notes, created_by, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
		)
	`

	conn := r.db.getConn(ctx)
	_, err := conn.Exec(ctx, q,
		m.ProductID,
		m.MutationType,
		m.Quantity,
		m.PreviousQty,
		m.NewQty,
		m.ReferenceID,
		m.ReferenceType,
		m.Notes,
		m.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert mutation: %w", err)
	}

	return nil
}

func (r *StockRepo) GetMutationHistory(ctx context.Context, productID string, limit, offset int) ([]*domain.StockMutation, error) {
	q := `
		SELECT id, product_id, mutation_type, quantity, previous_qty, new_qty,
			   reference_id, reference_type, notes, created_by, created_at
		FROM stock_mutations
		WHERE product_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	conn := r.db.getConn(ctx)
	rows, err := conn.Query(ctx, q, productID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query mutations: %w", err)
	}
	defer rows.Close()

	var mutations []*domain.StockMutation
	for rows.Next() {
		var m domain.StockMutation
		err := rows.Scan(
			&m.ID,
			&m.ProductID,
			&m.MutationType,
			&m.Quantity,
			&m.PreviousQty,
			&m.NewQty,
			&m.ReferenceID,
			&m.ReferenceType,
			&m.Notes,
			&m.CreatedBy,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan mutation: %w", err)
		}
		mutations = append(mutations, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return mutations, nil
}
