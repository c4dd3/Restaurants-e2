package repopg

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
)

func TestRestaurantRepoPgCRUD(t *testing.T) {
	pool := testPool(t)
	userRepo := NewUserRepoPg(pool)
	repo := NewRestaurantRepoPg(pool)
	ctx := context.Background()

	admin := seedAdminUser(t, userRepo)

	r := &domain.Restaurant{
		ID:       uuid.NewString(),
		Name:     "Soda La Tica",
		Address:  "Heredia, Centro",
		Phone:    "+506 2260-0001",
		AdminID:  admin.ID,
		Capacity: 25,
	}

	// Create
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() {
		t.Error("Create no llenó timestamps")
	}
	t.Cleanup(func() {
		// Sin cascada en la tabla: borramos directo con SQL.
		pool.Exec(ctx, "DELETE FROM restaurants WHERE id=$1", r.ID) //nolint:errcheck
	})

	// FindByID
	found, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != r.Name {
		t.Errorf("nombre esperado %q, obtenido %q", r.Name, found.Name)
	}
	if found.AdminID != admin.ID {
		t.Errorf("AdminID esperado %q, obtenido %q", admin.ID, found.AdminID)
	}

	// FindAll — al menos el que acabamos de crear debe aparecer
	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	found2 := false
	for _, x := range all {
		if x.ID == r.ID {
			found2 = true
			break
		}
	}
	if !found2 {
		t.Error("FindAll no devolvió el restaurante recién creado")
	}
}

func TestRestaurantRepoPgFindByIDNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewRestaurantRepoPg(pool)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestRestaurantRepoPgFindAllEmpty(t *testing.T) {
	pool := testPool(t)
	repo := NewRestaurantRepoPg(pool)
	ctx := context.Background()

	// No podemos garantizar BD vacía, pero sí que FindAll no retorna nil.
	list, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if list == nil {
		t.Error("FindAll retornó nil en vez de slice vacío")
	}
}
