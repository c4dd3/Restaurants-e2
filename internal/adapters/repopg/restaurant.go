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

var _ ports.RestaurantRepository = (*RestaurantRepoPg)(nil)

type RestaurantRepoPg struct {
	pool *pgxpool.Pool
}

func NewRestaurantRepoPg(pool *pgxpool.Pool) *RestaurantRepoPg {
	return &RestaurantRepoPg{pool: pool}
}

// Create inserta un restaurante. El ID debe venir generado desde la capa de servicio.
func (r *RestaurantRepoPg) Create(ctx context.Context, rest *domain.Restaurant) error {
	const q = `
		INSERT INTO restaurants (id, name, address, phone, description, admin_id, capacity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, q,
		rest.ID, rest.Name, rest.Address, rest.Phone,
		rest.Description, rest.AdminID, rest.Capacity,
	).Scan(&rest.CreatedAt, &rest.UpdatedAt)
	if err != nil {
		return pgErr(err)
	}
	return nil
}

func (r *RestaurantRepoPg) FindByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	const q = `
		SELECT id, name, address, phone, description, admin_id, capacity, created_at, updated_at
		FROM restaurants WHERE id = $1`

	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("query restaurant: %w", err)
	}
	rest, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Restaurant])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect restaurant: %w", err)
	}
	return &rest, nil
}

// FindAll retorna todos los restaurantes ordenados por fecha de creación descendente.
// Garantiza slice no-nil — retorna []domain.Restaurant{} si no hay registros.
func (r *RestaurantRepoPg) FindAll(ctx context.Context) ([]domain.Restaurant, error) {
	const q = `
		SELECT id, name, address, phone, description, admin_id, capacity, created_at, updated_at
		FROM restaurants
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query restaurants: %w", err)
	}
	restaurants, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Restaurant])
	if err != nil {
		return nil, fmt.Errorf("collect restaurants: %w", err)
	}
	// pgx v5 CollectRows siempre retorna un slice no-nil; este return es el caso base.
	return restaurants, nil
}
