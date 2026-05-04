package repomongo

import (
	"context"
	"testing"
	"time"

	"restaurants-e2/internal/domain"
)

// Acá me interesa validar la disponibilidad y el cambio de estado al cancelar.
func TestReservationRepoMongoCreateFindCancelAndAvailability(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	if err := repos.Restaurants.Create(ctx, &domain.Restaurant{ID: "rest-1", Name: "Soda TEC", Capacity: 10}); err != nil {
		t.Fatal(err)
	}
	confirmed := &domain.Reservation{ID: "res-confirmed", RestaurantID: "rest-1", UserID: "user-1", Date: time.Now().UTC().Add(30 * time.Minute), PartySize: 3, Status: domain.StatusConfirmed}
	if err := repos.Reservations.Create(ctx, confirmed); err != nil {
		t.Fatal(err)
	}

	available, err := repos.Reservations.CheckAvailability(ctx, "rest-1", 2)
	if err != nil {
		t.Fatal(err)
	}
	if available != 7 {
		t.Fatalf("disponibilidad esperada 7, obtuvo %d", available)
	}

	found, err := repos.Reservations.FindByID(ctx, "res-confirmed")
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || found.PartySize != 3 {
		t.Fatalf("reserva incorrecta: %#v", found)
	}

	if err := repos.Reservations.Cancel(ctx, "res-confirmed"); err != nil {
		t.Fatal(err)
	}
	cancelled, _ := repos.Reservations.FindByID(ctx, "res-confirmed")
	if cancelled.Status != domain.StatusCancelled {
		t.Fatalf("estado esperado cancelled, obtuvo %s", cancelled.Status)
	}
}
