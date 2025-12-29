package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/samsirama/go-wms-saas/internal/core/domain"
)

type ProductRepo struct {
	db *PostgresDB
}

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{
		db: NewPostgresDB(pool),
	}
}

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) error {
	q := `
		INSERT INTO products (id, sku, name, description, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	conn := r.db.getConn(ctx)
	row := conn.QueryRow(ctx, q, p.SKU, p.Name, p.Description)

	err := row.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}

	return nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	q := `
		SELECT id, sku, name, description, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	conn := r.db.getConn(ctx)
	row := conn.QueryRow(ctx, q, id)

	var p domain.Product
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan product: %w", err)
	}

	return &p, nil
}

func (r *ProductRepo) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	q := `
		SELECT id, sku, name, description, created_at, updated_at
		FROM products
		WHERE sku = $1
	`

	conn := r.db.getConn(ctx)
	row := conn.QueryRow(ctx, q, sku)

	var p domain.Product
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan product: %w", err)
	}

	return &p, nil
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) error {
	q := `
		UPDATE products
		SET sku = $1, name = $2, description = $3, updated_at = NOW()
		WHERE id = $4
	`

	conn := r.db.getConn(ctx)
	tag, err := conn.Exec(ctx, q, p.SKU, p.Name, p.Description, p.ID)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM products WHERE id = $1`

	conn := r.db.getConn(ctx)
	tag, err := conn.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ProductRepo) List(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	q := `
		SELECT id, sku, name, description, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	conn := r.db.getConn(ctx)
	rows, err := conn.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}
