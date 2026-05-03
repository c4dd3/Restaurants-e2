package repopg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

var _ ports.MenuRepository = (*MenuRepoPg)(nil)

type MenuRepoPg struct {
	pool *pgxpool.Pool
}

func NewMenuRepoPg(pool *pgxpool.Pool) *MenuRepoPg {
	return &MenuRepoPg{pool: pool}
}

// Create inserta el menú. Los productos asociados se crean por separado
// via ProductRepoPg. Si se quiere atomicidad menú+productos, el service
// debe usar una transacción de nivel superior.
func (r *MenuRepoPg) Create(ctx context.Context, m *domain.Menu) error {
	const q = `
		INSERT INTO menus (id, restaurant_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, q, m.ID, m.RestaurantID, m.Name, m.Description).
		Scan(&m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return pgErr(err)
	}
	return nil
}

// FindByID retorna el menú con sus productos cargados en dos queries.
func (r *MenuRepoPg) FindByID(ctx context.Context, id string) (*domain.Menu, error) {
	const q = `
		SELECT id, restaurant_id, name, description, created_at, updated_at
		FROM menus WHERE id = $1`

	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("query menu: %w", err)
	}
	m, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Menu])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect menu: %w", err)
	}

	m.Products, err = fetchProductsByMenuID(ctx, r.pool, id)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Update modifica name/description del menú y, si req.Products != nil,
// reemplaza TODOS los productos en una transacción atómica.
func (r *MenuRepoPg) Update(ctx context.Context, id string, req *domain.UpdateMenuRequest) (*domain.Menu, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op si ya se hizo Commit

	const updateQ = `
		UPDATE menus
		SET
			name        = COALESCE(NULLIF($2, ''), name),
			description = COALESCE(NULLIF($3, ''), description),
			updated_at  = NOW()
		WHERE id = $1
		RETURNING id, restaurant_id, name, description, created_at, updated_at`

	rows, err := tx.Query(ctx, updateQ, id, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("update menu: %w", err)
	}
	m, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Menu])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect menu: %w", err)
	}

	if req.Products != nil {
		if err := replaceProducts(ctx, tx, id, m.RestaurantID, req.Products); err != nil {
			return nil, err
		}
		m.Products, err = fetchProductsByMenuID(ctx, tx, id)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &m, nil
}

// Delete elimina el menú. El schema tiene ON DELETE CASCADE sobre products.menu_id,
// así que los productos del menú se eliminan automáticamente.
func (r *MenuRepoPg) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM menus WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete menu: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// querier es una interfaz mínima compartida por *pgxpool.Pool y pgx.Tx.
// Permite que fetchProductsByMenuID funcione dentro y fuera de una transacción.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// fetchProductsByMenuID carga los productos de un menú desde cualquier querier
// (pool o tx). Retorna slice vacío si no hay productos.
func fetchProductsByMenuID(ctx context.Context, q querier, menuID string) ([]domain.Product, error) {
	const query = `
		SELECT id, menu_id, restaurant_id, name, description, category, price, available
		FROM products
		WHERE menu_id = $1
		ORDER BY name`

	rows, err := q.Query(ctx, query, menuID)
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

// replaceProducts elimina todos los productos de un menú e inserta los nuevos,
// todo dentro de la transacción tx recibida.
func replaceProducts(ctx context.Context, tx pgx.Tx, menuID, restaurantID string, reqs []domain.ProductRequest) error {
	if _, err := tx.Exec(ctx, `DELETE FROM products WHERE menu_id = $1`, menuID); err != nil {
		return fmt.Errorf("delete products: %w", err)
	}

	const insertQ = `
		INSERT INTO products (id, menu_id, restaurant_id, name, description, category, price, available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	for _, p := range reqs {
		if _, err := tx.Exec(ctx, insertQ,
			uuid.New().String(), menuID, restaurantID,
			p.Name, p.Description, p.Category, p.Price, p.Available,
		); err != nil {
			return fmt.Errorf("insert product: %w", err)
		}
	}
	return nil
}
