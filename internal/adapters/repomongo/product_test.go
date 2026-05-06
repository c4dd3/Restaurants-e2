package repomongo

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

	"restaurants-e2/internal/domain"
)

// Además del CRUD, aquí pruebo consultas por ids, categoría y listado general.
func TestProductRepoMongoCRUDAndQueries(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	p := &domain.Product{
		ID:           "prod-1",
		MenuID:       "menu-1",
		RestaurantID: "rest-1",
		Name:         "Pizza",
		Category:     "pizzas",
		Price:        4500,
		Available:    true,
	}
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

	updated, err := repos.Products.FindByID(ctx, "prod-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated == nil || updated.Price != 5000 {
		t.Fatalf("precio no actualizado: %#v", updated)
	}

	if err := repos.Products.Delete(ctx, "prod-1"); err != nil {
		t.Fatal(err)
	}
}

func TestProductRepoMongoMoreBranches(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	p := &domain.Product{
		MenuID:       "menu-1",
		RestaurantID: "rest-1",
		Name:         "Té frío",
		Category:     "bebidas",
		Price:        1200,
		Available:    true,
	}
	if err := repos.Products.Create(ctx, p); err != nil {
		t.Fatal(err)
	}
	if p.ID == "" {
		t.Fatal("Create no generó id para producto")
	}

	// FindByID con id inexistente retorna ErrNotFound, no nil.
	_, err := repos.Products.FindByID(ctx, "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("esperaba ErrNotFound para producto inexistente, obtuvo %v", err)
	}

	byIDs, err := repos.Products.FindByIDs(ctx, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(byIDs) != 0 {
		t.Fatalf("FindByIDs vacío debería devolver 0, obtuvo %d", len(byIDs))
	}

	byCat, err := repos.Products.FindByCategory(ctx, "no-existe")
	if err != nil {
		t.Fatal(err)
	}
	if len(byCat) != 0 {
		t.Fatalf("categoría inexistente debería devolver 0, obtuvo %d", len(byCat))
	}

	// Update inexistente: UpdateOne sin MatchedCount check → no-op, sin error.
	err = repos.Products.Update(ctx, &domain.Product{
		ID:       "no-existe",
		Name:     "Nada",
		Category: "x",
	})
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("error inesperado al actualizar inexistente: %v", err)
	}

	// Confirmamos que no apareció un producto fantasma.
	_, err = repos.Products.FindByID(ctx, "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("esperaba ErrNotFound después de update inexistente, obtuvo %v", err)
	}

	// Delete inexistente: DeleteOne sin MatchedCount check → no-op, sin error.
	if err := repos.Products.Delete(ctx, "no-existe"); err != nil {
		t.Fatalf("esperaba nil al borrar producto inexistente, obtuvo %v", err)
	}
}

func TestNewProductRepository(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()

	if repos.Products == nil {
		t.Fatal("Products repo no debería ser nil")
	}
}

// TestProductRepoMongoValidationNoDB cubre las ramas de validación en product.go
// que retornan ANTES de tocar MongoDB. No requiere conexión real — las estructuras
// se crean directamente en el paquete (package repomongo).
func TestProductRepoMongoValidationNoDB(t *testing.T) {
	repo := &ProductRepoMongo{coll: nil} // coll nunca se alcanza en estas ramas
	ctx := context.Background()

	// Create con categoría vacía → error antes de InsertOne.
	if err := repo.Create(ctx, &domain.Product{ID: "x", Name: "Test", Category: ""}); err == nil {
		t.Error("Create: esperaba error por category vacía, obtuvo nil")
	}

	// Update con ID vacío → error antes de UpdateOne.
	if err := repo.Update(ctx, &domain.Product{Name: "Test", Category: "cat"}); err == nil {
		t.Error("Update: esperaba error por ID vacío, obtuvo nil")
	}

	// Update con categoría vacía → error antes de UpdateOne.
	if err := repo.Update(ctx, &domain.Product{ID: "x", Name: "Test", Category: ""}); err == nil {
		t.Error("Update: esperaba error por category vacía, obtuvo nil")
	}
}
