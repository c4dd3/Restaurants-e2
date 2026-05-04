package repomongo

import (
	"context"
	"testing"

	"restaurants-e2/internal/domain"
)

// Además del CRUD, aquí pruebo consultas por ids, categoría y listado general.
func TestProductRepoMongoCRUDAndQueries(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	p := &domain.Product{ID: "prod-1", MenuID: "menu-1", RestaurantID: "rest-1", Name: "Pizza", Category: "pizzas", Price: 4500, Available: true}
	if err := repos.Products.Create(ctx, p); err != nil {
		t.Fatal(err)
	}

	found, err := repos.Products.FindByID(ctx, "prod-1")
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || found.Name != "Pizza" {
		t.Fatalf("producto incorrecto: %#v", found)
	}

	byIDs, err := repos.Products.FindByIDs(ctx, []string{"prod-1", "no-existe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(byIDs) != 1 {
		t.Fatalf("FindByIDs esperaba 1, obtuvo %d", len(byIDs))
	}

	byCat, err := repos.Products.FindByCategory(ctx, "pizzas")
	if err != nil {
		t.Fatal(err)
	}
	if len(byCat) != 1 {
		t.Fatalf("FindByCategory esperaba 1, obtuvo %d", len(byCat))
	}

	all, err := repos.Products.FindAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("FindAll esperaba 1, obtuvo %d", len(all))
	}

	p.Price = 5000
	if err := repos.Products.Update(ctx, p); err != nil {
		t.Fatal(err)
	}
	updated, _ := repos.Products.FindByID(ctx, "prod-1")
	if updated.Price != 5000 {
		t.Fatalf("precio no actualizado: %#v", updated)
	}

	if err := repos.Products.Delete(ctx, "prod-1"); err != nil {
		t.Fatal(err)
	}
}
