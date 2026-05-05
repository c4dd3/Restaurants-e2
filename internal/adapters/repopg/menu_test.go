package repopg

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
)

func TestMenuRepoPgCRUD(t *testing.T) {
	pool := testPool(t)
	userRepo := NewUserRepoPg(pool)
	restRepo := NewRestaurantRepoPg(pool)
	repo := NewMenuRepoPg(pool)
	ctx := context.Background()

	admin := seedAdminUser(t, userRepo)
	rest := seedRestaurant(t, restRepo, admin.ID)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM restaurants WHERE id=$1", rest.ID) }) //nolint:errcheck

	m := &domain.Menu{
		ID:           uuid.NewString(),
		RestaurantID: rest.ID,
		Name:         "Menú Principal",
		Description:  "Comida típica",
	}

	// Create
	if err := repo.Create(ctx, m); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if m.CreatedAt.IsZero() {
		t.Error("Create no llenó timestamps")
	}
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM menus WHERE id=$1", m.ID) }) //nolint:errcheck

	// FindByID
	found, err := repo.FindByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != m.Name {
		t.Errorf("nombre esperado %q, obtenido %q", m.Name, found.Name)
	}
	// Products debe ser slice vacío, no nil
	if found.Products == nil {
		t.Error("FindByID retornó nil en Products")
	}

	// Update nombre
	updated, err := repo.Update(ctx, m.ID, &domain.UpdateMenuRequest{Name: "Carta Especial"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Carta Especial" {
		t.Errorf("nombre no actualizado: %q", updated.Name)
	}

	// Update con reemplazo de productos
	updated, err = repo.Update(ctx, m.ID, &domain.UpdateMenuRequest{
		Products: []domain.ProductRequest{
			{Name: "Gallo Pinto", Category: "plato fuerte", Price: 3500, Available: true},
		},
	})
	if err != nil {
		t.Fatalf("Update con productos: %v", err)
	}
	if len(updated.Products) != 1 {
		t.Errorf("esperado 1 producto tras update, obtenidos %d", len(updated.Products))
	}

	// Delete (cascada elimina productos automáticamente)
	if err := repo.Delete(ctx, m.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = repo.FindByID(ctx, m.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound después de Delete, obtenido %v", err)
	}
}

func TestMenuRepoPgFindByIDNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewMenuRepoPg(pool)

	_, err := repo.FindByID(context.Background(), uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestMenuRepoPgDeleteNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewMenuRepoPg(pool)

	err := repo.Delete(context.Background(), uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestMenuRepoPgUpdateNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewMenuRepoPg(pool)

	// Actualizar un menú inexistente retorna ErrNotFound (rama RETURNING vacío).
	_, err := repo.Update(context.Background(), uuid.NewString(), &domain.UpdateMenuRequest{Name: "Nada"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound al actualizar menú inexistente, obtenido %v", err)
	}
}

func TestMenuRepoPgUpdateEmptyProducts(t *testing.T) {
	pool := testPool(t)
	userRepo := NewUserRepoPg(pool)
	restRepo := NewRestaurantRepoPg(pool)
	repo := NewMenuRepoPg(pool)
	ctx := context.Background()

	admin := seedAdminUser(t, userRepo)
	rest := seedRestaurant(t, restRepo, admin.ID)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM restaurants WHERE id=$1", rest.ID) }) //nolint:errcheck

	m := &domain.Menu{ID: uuid.NewString(), RestaurantID: rest.ID, Name: "Menú Vacío", Description: "sin items"}
	if err := repo.Create(ctx, m); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM menus WHERE id=$1", m.ID) }) //nolint:errcheck

	// Products: []ProductRequest{} (slice vacío, no nil) → replaceProducts borra todo y no inserta nada.
	updated, err := repo.Update(ctx, m.ID, &domain.UpdateMenuRequest{
		Name:     "Actualizado",
		Products: []domain.ProductRequest{},
	})
	if err != nil {
		t.Fatalf("Update con slice vacío: %v", err)
	}
	if updated.Name != "Actualizado" {
		t.Errorf("nombre no actualizado: %q", updated.Name)
	}
	if len(updated.Products) != 0 {
		t.Errorf("esperado 0 productos, obtenidos %d", len(updated.Products))
	}
}
