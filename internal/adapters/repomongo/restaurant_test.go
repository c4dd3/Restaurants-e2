package repomongo

import (
	"context"
	"errors"
	"testing"

	"restaurants-e2/internal/domain"
)

// Caso simple para restaurantes: crear, buscar por id y listar.
func TestRestaurantRepoMongoCreateFindAndList(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	r := &domain.Restaurant{ID: "rest-1", Name: "Soda TEC", Address: "Cartago", Phone: "8888-8888", AdminID: "admin-1", Capacity: 20}
	if err := repos.Restaurants.Create(ctx, r); err != nil {
		t.Fatal(err)
	}

	found, err := repos.Restaurants.FindByID(ctx, "rest-1")
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || found.Name != "Soda TEC" {
		t.Fatalf("restaurante incorrecto: %#v", found)
	}

	all, err := repos.Restaurants.FindAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("se esperaba 1 restaurante, obtuvo %d", len(all))
	}
}

// Revisa campos automáticos y el caso donde no se encuentra el restaurante.
func TestRestaurantRepoMongoDefaultsAndMissing(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	r := &domain.Restaurant{Name: "Rest sin ID", Address: "Cartago", Phone: "2222-2222", AdminID: "admin-1", Capacity: 12}
	if err := repos.Restaurants.Create(ctx, r); err != nil {
		t.Fatal(err)
	}
	if r.ID == "" || r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() {
		t.Fatalf("no llenó campos automáticos: %#v", r)
	}

	missing, err := repos.Restaurants.FindByID(ctx, "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("esperaba ErrNotFound, obtuvo restaurant=%#v err=%v", missing, err)
	}
}
