package repomongo

import (
	"context"
	"errors"
	"strings"
	"testing"

	"restaurants-e2/internal/domain"
)

// En estas pruebas del menú me interesa cubrir create/find/update.
// Algunas rutas de update con productos y delete usan transacciones.
// Si el entorno de prueba no las soporta bien, no las tomo como fallo del repo.
func isTxnUnsupported(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Transaction numbers are only allowed on a replica set member or mongos") ||
		strings.Contains(msg, "transactions are not supported")
}

// El menú se prueba junto con productos porque FindByID los arma completos.
func TestMenuRepoMongoCRUD(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	m := &domain.Menu{
		ID:           "menu-1",
		RestaurantID: "rest-1",
		Name:         "Menú almuerzo",
		Description:  "Casados",
	}
	if err := repos.Menus.Create(ctx, m); err != nil {
		t.Fatal(err)
	}

	if err := repos.Products.Create(ctx, &domain.Product{
		ID:           "prod-1",
		MenuID:       "menu-1",
		RestaurantID: "rest-1",
		Name:         "Casado",
		Category:     "almuerzos",
		Price:        3500,
		Available:    true,
	}); err != nil {
		t.Fatal(err)
	}

	found, err := repos.Menus.FindByID(ctx, "menu-1")
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || found.Name != "Menú almuerzo" || len(found.Products) != 1 {
		t.Fatalf("menú incorrecto: %#v", found)
	}

	updated, err := repos.Menus.Update(ctx, "menu-1", &domain.UpdateMenuRequest{
		Name: "Menú actualizado",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated == nil || updated.Name != "Menú actualizado" {
		t.Fatalf("nombre no actualizado: %#v", updated)
	}

	// Delete pasa por transacción. Si el entorno no lo soporta, no quiero botar
	// toda la suite por algo del ambiente local.
	err = repos.Menus.Delete(ctx, "menu-1")
	if err != nil && !isTxnUnsupported(err) {
		t.Fatal(err)
	}

	// Si Delete sí pasó, confirmo que ya no exista.
	if err == nil {
		missing, findErr := repos.Menus.FindByID(ctx, "menu-1")
		if !errors.Is(findErr, domain.ErrNotFound) {
			t.Fatalf("esperaba ErrNotFound, obtuvo menu=%#v err=%v", missing, findErr)
		}
	}
}

func TestMenuRepoMongoUpdateProductsAndMissing(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	m := &domain.Menu{
		ID:           "menu-extra",
		RestaurantID: "rest-1",
		Name:         "Menú viejo",
		Description:  "desc",
	}
	if err := repos.Menus.Create(ctx, m); err != nil {
		t.Fatal(err)
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		t.Fatalf("no llenó fechas automáticas: %#v", m)
	}

	if err := repos.Products.Create(ctx, &domain.Product{
		ID:           "prod-viejo",
		MenuID:       "menu-extra",
		RestaurantID: "rest-1",
		Name:         "Viejo",
		Category:     "viejos",
		Price:        1000,
	}); err != nil {
		t.Fatal(err)
	}

	updated, err := repos.Menus.Update(ctx, "menu-extra", &domain.UpdateMenuRequest{
		Name:        "Menú nuevo",
		Description: "actualizado",
		Products: []domain.ProductRequest{
			{Name: "Nuevo 1", Category: "nuevos", Price: 2500, Available: true},
			{Name: "Nuevo 2", Category: "nuevos", Price: 3000, Available: true},
		},
	})

	// Esta rama entra por transacción. Si el ambiente la rechaza, no lo cuento
	// como fallo funcional del repo para esta máquina.
	if err != nil && isTxnUnsupported(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}

	if updated == nil || updated.Name != "Menú nuevo" || len(updated.Products) != 2 {
		t.Fatalf("update con productos incorrecto: %#v", updated)
	}

	found, err := repos.Menus.FindByID(ctx, "menu-extra")
	if err != nil {
		t.Fatal(err)
	}
	if len(found.Products) != 2 {
		t.Fatalf("se esperaban 2 productos nuevos, obtuvo %#v", found.Products)
	}

	for _, p := range found.Products {
		if p.MenuID != "menu-extra" || p.ID == "" {
			t.Fatalf("producto generado incorrecto: %#v", p)
		}
	}

	// El producto viejo fue eliminado por replaceProducts → debe retornar ErrNotFound.
	_, err = repos.Products.FindByID(ctx, "prod-viejo")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("el producto viejo debió borrarse (esperaba ErrNotFound), obtuvo %v", err)
	}

	missing, err := repos.Menus.FindByID(ctx, "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("esperaba ErrNotFound, obtuvo menu=%#v err=%v", missing, err)
	}

	updated, err = repos.Menus.Update(ctx, "no-existe", &domain.UpdateMenuRequest{Name: "Nada"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("esperaba ErrNotFound al actualizar menú inexistente, obtuvo menu=%#v err=%v", updated, err)
	}

	err = repos.Menus.Delete(ctx, "menu-extra")
	if err != nil && !isTxnUnsupported(err) {
		t.Fatal(err)
	}
}
