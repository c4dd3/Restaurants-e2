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

var _ ports.OrderRepository = (*OrderRepoPg)(nil)

type OrderRepoPg struct {
	pool *pgxpool.Pool
}

func NewOrderRepoPg(pool *pgxpool.Pool) *OrderRepoPg {
	return &OrderRepoPg{pool: pool}
}

// Create persiste la orden y sus items en una sola transacción.
// Los items se insertan con pgx.Batch — una sola llamada al servidor
// para N inserts, en vez de N round-trips separados.
//
// El precio de cada item se congela al momento de crear la orden:
// o.Items[i].Price debe venir calculado por la capa de servicio
// (quien consulta el precio actual del producto antes de llamar acá).
func (r *OrderRepoPg) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insertar la orden.
	const orderQ = `
		INSERT INTO orders (id, user_id, restaurant_id, reservation_id, total, status, pickup, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at`

	err = tx.QueryRow(ctx, orderQ,
		o.ID, o.UserID, o.RestaurantID, o.ReservationID,
		o.Total, o.Status, o.Pickup,
	).Scan(&o.CreatedAt)
	if err != nil {
		return pgErr(err)
	}

	// 2. Insertar los items en batch (un solo round-trip al servidor).
	if len(o.Items) > 0 {
		if err := insertOrderItems(ctx, tx, o.ID, o.Items); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// FindByID retorna la orden con sus items cargados en dos queries.
func (r *OrderRepoPg) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	const q = `
		SELECT id, user_id, restaurant_id, reservation_id, total, status, pickup, created_at
		FROM orders WHERE id = $1`

	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("query order: %w", err)
	}
	o, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Order])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect order: %w", err)
	}

	o.Items, err = fetchOrderItems(ctx, r.pool, id)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// insertOrderItems encola todos los inserts en un pgx.Batch y los envía al
// servidor en una sola llamada. Cada item recibe un UUID nuevo.
func insertOrderItems(ctx context.Context, tx pgx.Tx, orderID string, items []domain.OrderItem) error {
	const q = `
		INSERT INTO order_items (id, order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4, $5)`

	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue(q, uuid.New().String(), orderID, item.ProductID, item.Quantity, item.Price)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range items {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("insert order_item: %w", err)
		}
	}
	return nil
}

// fetchOrderItems carga los items de una orden. Retorna slice vacío si no hay items.
func fetchOrderItems(ctx context.Context, pool *pgxpool.Pool, orderID string) ([]domain.OrderItem, error) {
	const q = `
		SELECT id, order_id, product_id, quantity, price
		FROM order_items
		WHERE order_id = $1`

	rows, err := pool.Query(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("query order_items: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.OrderItem])
	if err != nil {
		return nil, fmt.Errorf("collect order_items: %w", err)
	}
	// pgx v5 CollectRows nunca retorna nil — retorna slice vacío si no hay filas.
	return items, nil
}
