package repopg

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
)

func TestProductRepoPgCRUD(t *testing.T) {
	pool := testPool(t)
	userRepo := NewUserRepoPg(pool)
	restRepo := NewRestaurantRepoPg(pool)
	repo := NewProductRepoPg(pool)
	ctx := context.Background()

	// Fixtures: user → restaurant → menu (foreign keys requeridas por products)
	admin := seedAdminUser(t, userRepo)
	rest := seedRestaurant(t, restRepo, admin.ID)

	menuID := uuid.NewString()
	if _, err := pool.Exec(ctx,
		"INSERT INTO menus (id, restaurant_id, name, description, created_at, updated_at) VALUES ($1,$2,$3,$4,NOW(),NOW())",
		menuID, rest.ID, "Menú Test", "desc",
	); err != nil {
		t.Fatalf("insertar menú fixture: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM products WHERE menu_id=$1", menuID)   //nolint:errcheck
		pool.Exec(ctx, "DELETE FROM menus WHERE id=$1", menuID)           //nolint:errcheck
		pool.Exec(ctx, "DELETE FROM restaurants WHERE id=$1", rest.ID)    //nolint:errcheck
	})

	p := &domain.Product{
		ID:           uuid.NewString(),
		MenuID:       menuID,
		RestaurantID: rest.ID,
		Name:         "Gallo Pinto",
		Description:  "Desayuno típico",
		Category:     "plato fuerte",
		Price:        3500,
		Available:    true,
	}

	// Create
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// FindByID
	found, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != p.Name || found.Price != p.Price {
		t.Errorf("datos incorrectos: %+v", found)
	}

	// FindByIDs
	byIDs, err := repo.FindByIDs(ctx, []string{p.ID})
	if err != nil {
		t.Fatalf("FindByIDs: %v", err)
	}
	if len(byIDs) != 1 {
		t.Errorf("FindByIDs esperaba 1, obtuvo %d", len(byIDs))
	}

	// FindByCategory
	byCat, err := repo.FindByCategory(ctx, "plato fuerte")
	if err != nil {
		t.Fatalf("FindByCategory: %v", err)
	}
	foundInCat := false
	for _, x := range byCat {
		if x.ID == p.ID {
			foundInCat = true
			break
		}
	}
	if !foundInCat {
		t.Error("producto no apareció en FindByCategory")
	}

	// Update
	p.Name = "Gallo Pinto Mejorado"
	p.Price = 4000
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("Update: %v", err)
	}
	updated, _ := repo.FindByID(ctx, p.ID)
	if updated.Name != "Gallo Pinto Mejorado" || updated.Price != 4000 {
		t.Errorf("Update no persistió cambios: %+v", updated)
	}

	// Delete
	if err := repo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = repo.FindByID(ctx, p.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound después de Delete, obtenido %v", err)
	}
}

func TestProductRepoPgFindByIDsEmpty(t *testing.T) {
	pool := testPool(t)
	repo := NewProductRepoPg(pool)
	ctx := context.Background()

	list, err := repo.FindByIDs(ctx, []string{})
	if err != nil {
		t.Fatalf("FindByIDs(empty): %v", err)
	}
	if len(list) != 0 {
		t.Errorf("esperaba slice vacío, obtuvo %d elementos", len(list))
	}
}

func TestProductRepoPgUpdateNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewProductRepoPg(pool)
	ctx := context.Background()

	err := repo.Update(ctx, &domain.Product{ID: uuid.NewString(), Name: "X", Category: "postre", Price: 1000})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestProductRepoPgDeleteNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewProductRepoPg(pool)
	ctx := context.Background()

	err := repo.Delete(ctx, uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

// TestProductRepoPgFindAll verifica que FindAll retorna al menos el producto
// insertado en el CRUD test, y retorna slice no-nil si hay registros.
func TestProductRepoPgFindAll(t *testing.T) {
	pool := testPool(t)
	userRepo := NewUserRepoPg(pool)
	restRepo := NewRestaurantRepoPg(pool)
	repo := NewProductRepoPg(pool)
	ctx := context.Background()

	admin := seedAdminUser(t, userRepo)
	rest := seedRestaurant(t, restRepo, admin.ID)

	menuID := uuid.NewString()
	if _, err := pool.Exec(ctx,
		"INSERT INTO menus (id, restaurant_id, name, description, created_at, updated_at) VALUES ($1,$2,$3,$4,NOW(),NOW())",
		menuID, rest.ID, "Menú FindAll", "",
	); err != nil {
		t.Fatalf("insertar menú fixture: %v", err)
	}

	p := &domain.Product{
		ID:           uuid.NewString(),
		MenuID:       menuID,
		RestaurantID: rest.ID,
		Name:         "Producto FindAll Test",
		Category:     "postre",
		Price:        2000,
		Available:    true,
	}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM products WHERE id=$1", p.ID)           //nolint:errcheck
		pool.Exec(ctx, "DELETE FROM menus WHERE id=$1", menuID)            //nolint:errcheck
		pool.Exec(ctx, "DELETE FROM restaurants WHERE id=$1", rest.ID)     //nolint:errcheck
	})

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if all == nil {
		t.Fatal("FindAll retornó nil")
	}
	found := false
	for _, x := range all {
		if x.ID == p.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("producto %q no apareció en FindAll", p.ID)
	}
}
