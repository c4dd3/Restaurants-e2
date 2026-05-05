package repopg

// helpers_test.go — fixtures compartidos entre todos los _test.go del paquete.
// testPool está definido en user_test.go y disponible aquí por ser el mismo package.

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
)

// seedAdminUser crea un usuario admin en la BD para tests que necesiten admin_id.
func seedAdminUser(t *testing.T, repo *UserRepoPg) *domain.User {
	t.Helper()
	ctx := context.Background()
	u := &domain.User{
		ID:       uuid.NewString(),
		Name:     "Admin Test",
		Email:    "admin-" + uuid.NewString()[:8] + "@test.com",
		Password: "hashed",
		Role:     domain.RoleAdmin,
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("seedAdminUser: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })
	return u
}

// seedRestaurant crea un restaurante en la BD que depende de un admin existente.
func seedRestaurant(t *testing.T, repo *RestaurantRepoPg, adminID string) *domain.Restaurant {
	t.Helper()
	ctx := context.Background()
	r := &domain.Restaurant{
		ID:       uuid.NewString(),
		Name:     "Restaurante Test " + uuid.NewString()[:6],
		Address:  "San José, Costa Rica",
		Phone:    "+506 2222-0000",
		AdminID:  adminID,
		Capacity: 30,
	}
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("seedRestaurant: %v", err)
	}
	return r
}
