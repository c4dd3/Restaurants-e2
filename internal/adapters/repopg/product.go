package repopg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

var _ ports.ProductRepository = (*ProductRepoPg)(nil)

type ProductRepoPg struct {
	pool *pgxpool.Pool
}

func NewProductRepoPg(pool *pgxpool.Pool) *ProductRepoPg {
	return &ProductRepoPg{pool: pool}
}

func (r *ProductRepoPg) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	const q = `
		SELECT id, menu_id, restaurant_id, name, description, category, price, available
		FROM products WHERE id = $1`

	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}
	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Product])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect product: %w", err)
	}
	return &p, nil
}

// FindByCategory usa el índice idx_products_category definido en init.sql.
// Retorna slice vacío si no hay productos en esa categoría.
func (r *ProductRepoPg) FindByCategory(ctx context.Context, category string) ([]domain.Product, error) {
	const q = `
		SELECT id, menu_id, restaurant_id, name, description, category, price, available
		FROM products
		WHERE category = $1
		ORDER BY name`

	return collectProducts(ctx, r.pool, q, category)
}

// FindAll retorna todos los productos ordenados por nombre.
// Si el catálogo crece considera agregar paginación por keyset (WHERE id > $cursor LIMIT n).
func (r *ProductRepoPg) FindAll(ctx context.Context) ([]domain.Product, error) {
	const q = `
		SELECT id, menu_id, restaurant_id, name, description, category, price, available
		FROM products
		ORDER BY name`

	return collectProducts(ctx, r.pool, q)
}

// Create inserta un producto. La descripción se persiste tal cual — vacía si el caller
// no la provee. El getter domain.Product.DescriptionOrDefault() aplica el fallback al leer.
func (r *ProductRepoPg) Create(ctx context.Context, p *domain.Product) error {
	const q = `
		INSERT INTO products (id, menu_id, restaurant_id, name, description, category, price, available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, q,
		p.ID, p.MenuID, p.RestaurantID,
		p.Name, p.Description, p.Category, p.Price, p.Available,
	)
	if err != nil {
		return pgErr(err)
	}
	return nil
}

// Update reemplaza todos los campos editables del producto.
func (r *ProductRepoPg) Update(ctx context.Context, p *domain.Product) error {
	const q = `
		UPDATE products
		SET name=$2, description=$3, category=$4, price=$5, available=$6
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q,
		p.ID, p.Name, p.Description, p.Category, p.Price, p.Available,
	)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductRepoPg) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM products WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// collectProducts ejecuta q con args y retorna un slice de productos.
// Centraliza el manejo de rows para FindAll y FindByCategory.
func collectProducts(ctx context.Context, pool *pgxpool.Pool, q string, args ...any) ([]domain.Product, error) {
	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Product])
	if err != nil {
		return nil, fmt.Errorf("collect products: %w", err)
	}
	if products == nil {
		return []domain.Product{}, nil
	}
	return products, nil
}
