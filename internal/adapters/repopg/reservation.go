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

var _ ports.ReservationRepository = (*ReservationRepoPg)(nil)

type ReservationRepoPg struct {
	pool *pgxpool.Pool
}

func NewReservationRepoPg(pool *pgxpool.Pool) *ReservationRepoPg {
	return &ReservationRepoPg{pool: pool}
}

// Create inserta una reserva con status inicial "pending".
// Si el INSERT viola el constraint de exclusión (solapamiento de reservas en la misma
// ventana de 2h), pgErr lo traduce a domain.ErrConflict.
func (r *ReservationRepoPg) Create(ctx context.Context, res *domain.Reservation) error {
	const q = `
		INSERT INTO reservations (id, restaurant_id, user_id, date, party_size, status, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, q,
		res.ID, res.RestaurantID, res.UserID,
		res.Date, res.PartySize, res.Status, res.Notes,
	).Scan(&res.CreatedAt)
	if err != nil {
		return pgErr(err)
	}
	return nil
}

func (r *ReservationRepoPg) FindByID(ctx context.Context, id string) (*domain.Reservation, error) {
	const q = `
		SELECT id, restaurant_id, user_id, date, party_size, status, notes, created_at
		FROM reservations WHERE id = $1`

	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("query reservation: %w", err)
	}
	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Reservation])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect reservation: %w", err)
	}
	return &res, nil
}

// Cancel marca una reserva como cancelada. Si el id no existe o ya estaba
// cancelada, retorna domain.ErrNotFound — el service puede ignorarlo o
// exponerlo según el caso de uso.
func (r *ReservationRepoPg) Cancel(ctx context.Context, id string) error {
	const q = `
		UPDATE reservations
		SET status = $2
		WHERE id = $1 AND status != $2`

	tag, err := r.pool.Exec(ctx, q, id, domain.StatusCancelled)
	if err != nil {
		return fmt.Errorf("cancel reservation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CheckAvailability calcula cuántos asientos quedan disponibles en el
// restaurante para la próxima ventana de 2 horas.
//
// Retorna el número de asientos disponibles. Si es menor que partySize,
// el service debe retornar ErrConflict — la decisión de negocio vive allá,
// no acá.
//
// Dos queries separadas en vez de un JOIN para mantener la legibilidad;
// el costo es mínimo dado el volumen esperado.
func (r *ReservationRepoPg) CheckAvailability(ctx context.Context, restaurantID string, partySize int) (int, error) {
	// 1. Capacidad total del restaurante.
	var capacity int
	err := r.pool.QueryRow(ctx,
		`SELECT capacity FROM restaurants WHERE id = $1`,
		restaurantID,
	).Scan(&capacity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrNotFound
		}
		return 0, fmt.Errorf("query capacity: %w", err)
	}

	// 2. Asientos ya ocupados por reservas confirmadas en la ventana de 2 horas.
	var occupied int
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(party_size), 0)
		FROM reservations
		WHERE restaurant_id = $1
		  AND status        = $2
		  AND date         >= NOW()
		  AND date          < NOW() + INTERVAL '2 hours'`,
		restaurantID, domain.StatusConfirmed,
	).Scan(&occupied)
	if err != nil {
		return 0, fmt.Errorf("query occupied seats: %w", err)
	}

	return capacity - occupied, nil
}
