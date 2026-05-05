package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"restaurants-e2/internal/domain"
)

func newReservationSvc(res *mockReservationRepo, rests *mockRestaurantRepo, cache *mockCache) *ReservationService {
	return NewReservationService(res, rests, cache)
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestReservationServiceCreateHappyPath(t *testing.T) {
	rests := newMockRestaurantRepo()
	rests.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1", Capacity: 50}
	res := newMockReservationRepo(30) // 30 asientos disponibles
	cache := newMockCache()
	svc := newReservationSvc(res, rests, cache)

	r, err := svc.Create(context.Background(), "user-1", domain.CreateReservationRequest{
		RestaurantID: "rest-1",
		Date:         time.Now().Add(24 * time.Hour),
		PartySize:    4,
	})

	if err != nil {
		t.Fatalf("Create inesperado: %v", err)
	}
	if r.ID == "" {
		t.Fatal("reserva sin ID")
	}
	if r.UserID != "user-1" {
		t.Errorf("UserID esperado user-1, obtenido %q", r.UserID)
	}
	if r.Status != domain.StatusPending {
		t.Errorf("estado esperado pending, obtenido %q", r.Status)
	}
	if len(cache.deletedKeys) == 0 {
		t.Error("Create no invalidó la cache")
	}
}

func TestReservationServiceCreateRestaurantNotFound(t *testing.T) {
	svc := newReservationSvc(newMockReservationRepo(10), newMockRestaurantRepo(), newMockCache())

	_, err := svc.Create(context.Background(), "user-1", domain.CreateReservationRequest{
		RestaurantID: "no-existe",
		Date:         time.Now().Add(time.Hour),
		PartySize:    2,
	})

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("esperado ErrValidation, obtenido %v", err)
	}
}

func TestReservationServiceCreateNoAvailability(t *testing.T) {
	rests := newMockRestaurantRepo()
	rests.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1", Capacity: 10}
	// Solo 1 asiento disponible, grupo de 5
	svc := newReservationSvc(newMockReservationRepo(1), rests, newMockCache())

	_, err := svc.Create(context.Background(), "user-1", domain.CreateReservationRequest{
		RestaurantID: "rest-1",
		Date:         time.Now().Add(time.Hour),
		PartySize:    5,
	})

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("esperado ErrConflict por falta de disponibilidad, obtenido %v", err)
	}
}

// ── Cancel ────────────────────────────────────────────────────────────────────

func TestReservationServiceCancelOwner(t *testing.T) {
	res := newMockReservationRepo(10)
	res.reservations["res-1"] = &domain.Reservation{
		ID: "res-1", UserID: "user-1", RestaurantID: "rest-1", Status: domain.StatusPending,
	}
	cache := newMockCache()
	svc := newReservationSvc(res, newMockRestaurantRepo(), cache)

	if err := svc.Cancel(context.Background(), "user-1", "res-1"); err != nil {
		t.Fatalf("Cancel inesperado: %v", err)
	}
	if res.reservations["res-1"].Status != domain.StatusCancelled {
		t.Error("reserva no cambió a cancelled")
	}
	if len(cache.deletedKeys) == 0 {
		t.Error("Cancel no invalidó la cache")
	}
}

func TestReservationServiceCancelForbiddenOtherUser(t *testing.T) {
	res := newMockReservationRepo(10)
	res.reservations["res-1"] = &domain.Reservation{
		ID: "res-1", UserID: "user-1", RestaurantID: "rest-1", Status: domain.StatusPending,
	}
	svc := newReservationSvc(res, newMockRestaurantRepo(), newMockCache())

	err := svc.Cancel(context.Background(), "user-2", "res-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

func TestReservationServiceCancelNotFound(t *testing.T) {
	svc := newReservationSvc(newMockReservationRepo(10), newMockRestaurantRepo(), newMockCache())

	err := svc.Cancel(context.Background(), "user-1", "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestReservationServiceCreateRestaurantUnexpectedError(t *testing.T) {
	// Cubre la rama: FindByID falla con un error que NO es ErrNotFound (línea 47).
	rests := newMockRestaurantRepo()
	rests.findByIDErr = errors.New("db connection lost")
	svc := newReservationSvc(newMockReservationRepo(10), rests, newMockCache())

	_, err := svc.Create(context.Background(), "user-1", domain.CreateReservationRequest{
		RestaurantID: "rest-1",
		Date:         time.Now().Add(time.Hour),
		PartySize:    2,
	})
	if err == nil || errors.Is(err, domain.ErrValidation) {
		t.Errorf("esperado error inesperado propagado, obtenido %v", err)
	}
}

func TestReservationServiceCreateCheckAvailabilityError(t *testing.T) {
	// Cubre la rama: CheckAvailability falla (línea 53).
	rests := newMockRestaurantRepo()
	rests.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1", Capacity: 50}
	res := newMockReservationRepo(10)
	res.checkAvailErr = errors.New("availability check failed")
	svc := newReservationSvc(res, rests, newMockCache())

	_, err := svc.Create(context.Background(), "user-1", domain.CreateReservationRequest{
		RestaurantID: "rest-1",
		Date:         time.Now().Add(time.Hour),
		PartySize:    2,
	})
	if err == nil {
		t.Error("esperado error de CheckAvailability propagado, obtenido nil")
	}
}
