package repomongo

import (
	"context"
	"testing"

	"restaurants-e2/internal/domain"
)

// El menú se prueba junto con productos porque FindByID los arma completos.
func TestMenuRepoMongoCRUD(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	m := &domain.Menu{ID: "menu-1", RestaurantID: "rest-1", Name: "Menú almuerzo", Description: "Casados"}
	if err := repos.Menus.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	if err := repos.Products.Create(ctx, &domain.Product{ID: "prod-1", MenuID: "menu-1", RestaurantID: "rest-1", Name: "Casado", Category: "almuerzos", Price: 3500, Available: true}); err != nil {
		t.Fatal(err)
	}

	found, err := repos.Menus.FindByID(ctx, "menu-1")
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || found.Name != "Menú almuerzo" || len(found.Products) != 1 {
		t.Fatalf("menú incorrecto: %#v", found)
	}

	updated, err := repos.Menus.Update(ctx, "menu-1", &domain.UpdateMenuRequest{Name: "Menú actualizado"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Menú actualizado" {
		t.Fatalf("nombre no actualizado: %#v", updated)
	}

	if err := repos.Menus.Delete(ctx, "menu-1"); err != nil {
		t.Fatal(err)
	}
	missing, err := repos.Menus.FindByID(ctx, "menu-1")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Fatalf("se esperaba menú eliminado, obtuvo %#v", missing)
	}
}
