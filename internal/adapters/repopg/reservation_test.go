package repopg

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
)

func TestReservationRepoPgCRUD(t *testing.T) {
	pool := testPool(t)
	userRepo := NewUserRepoPg(pool)
	restRepo := NewRestaurantRepoPg(pool)
	repo := NewReservationRepoPg(pool)
	ctx := context.Background()

	admin := seedAdminUser(t, userRepo)
	rest := seedRestaurant(t, restRepo, admin.ID)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM restaurants WHERE id=$1", rest.ID) }) //nolint:errcheck

	res := &domain.Reservation{
		ID:           uuid.NewString(),
		RestaurantID: rest.ID,
		UserID:       admin.ID,
		Date:         time.Now().UTC().Add(24 * time.Hour),
		PartySize:    4,
		Status:       domain.StatusPending,
		Notes:        "Mesa cerca de la ventana",
	}

	// Create
	if err := repo.Create(ctx, res); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.CreatedAt.IsZero() {
		t.Error("Create no llenó created_at")
	}
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM reservations WHERE id=$1", res.ID) }) //nolint:errcheck

	// FindByID
	found, err := repo.FindByID(ctx, res.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.PartySize != res.PartySize {
		t.Errorf("party_size esperado %d, obtenido %d", res.PartySize, found.PartySize)
	}
	if found.Status != domain.StatusPending {
		t.Errorf("status esperado pending, obtenido %q", found.Status)
	}
	if found.Notes != res.Notes {
		t.Errorf("notes esperado %q, obtenido %q", res.Notes, found.Notes)
	}

	// Cancel
	if err := repo.Cancel(ctx, res.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	cancelled, _ := repo.FindByID(ctx, res.ID)
	if cancelled.Status != domain.StatusCancelled {
		t.Errorf("status esperado cancelled, obtenido %q", cancelled.Status)
	}

	// Cancelar de nuevo → ErrNotFound (ya está cancelada, WHERE status != cancelled da 0 rows)
	if err := repo.Cancel(ctx, res.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("cancelar dos veces: esperado ErrNotFound, obtenido %v", err)
	}
}

func TestReservationRepoPgFindByIDNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewReservationRepoPg(pool)

	_, err := repo.FindByID(context.Background(), uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestReservationRepoPgCheckAvailability(t *testing.T) {
	pool := testPool(t)
	userRepo := NewUserRepoPg(pool)
	restRepo := NewRestaurantRepoPg(pool)
	repo := NewReservationRepoPg(pool)
	ctx := context.Background()

	admin := seedAdminUser(t, userRepo)
	rest := seedRestaurant(t, restRepo, admin.ID) // capacity = 30
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM restaurants WHERE id=$1", rest.ID) }) //nolint:errcheck

	// Sin reservas activas, disponibilidad = capacidad total
	available, err := repo.CheckAvailability(ctx, rest.ID, 5)
	if err != nil {
		t.Fatalf("CheckAvailability: %v", err)
	}
	if available != rest.Capacity {
		t.Errorf("disponibilidad esperada %d, obtenida %d", rest.Capacity, available)
	}

	// Restaurante inexistente → ErrNotFound
	_, err = repo.CheckAvailability(ctx, uuid.NewString(), 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound para restaurante inexistente, obtenido %v", err)
	}
}
