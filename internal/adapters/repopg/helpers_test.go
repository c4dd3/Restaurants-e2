package repopg

// helpers_test.go — fixtures compartidos entre todos los _test.go del paquete.
// testPool está definido en user_test.go y disponible aquí por ser el mismo package.

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

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

// ── pgErr unit tests (sin BD) ─────────────────────────────────────────────────

// TestPgErrNonPgError cubre la rama donde pgErr recibe un error que NO es
// pgconn.PgError: el if errors.As(...) es false y se ejecuta el return fmt.Errorf.
func TestPgErrNonPgError(t *testing.T) {
	err := pgErr(errors.New("error genérico"))
	if err == nil {
		t.Fatal("pgErr debería retornar error")
	}
	if errors.Is(err, domain.ErrConflict) {
		t.Fatal("error genérico no debería mapearse a ErrConflict")
	}
}

// TestPgErrUnknownPgCode cubre la rama donde pgErr recibe un pgconn.PgError
// con un código desconocido: el switch no tiene match y cae al return fmt.Errorf.
func TestPgErrUnknownPgCode(t *testing.T) {
	pgcErr := &pgconn.PgError{Code: "99999"} // código inexistente en el switch
	err := pgErr(pgcErr)
	if err == nil {
		t.Fatal("pgErr debería retornar error para código desconocido")
	}
	if errors.Is(err, domain.ErrConflict) {
		t.Fatalf("código desconocido no debería mapearse a ErrConflict, obtuvo: %v", err)
	}
}

// TestPgErrInvalidTextRepr cubre la rama 22P02 de pgErr:
// valor no casteable al tipo de columna → domain.ErrNotFound.
func TestPgErrInvalidTextRepr(t *testing.T) {
	pgcErr := &pgconn.PgError{Code: "22P02"}
	err := pgErr(pgcErr)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("22P02 debería mapearse a ErrNotFound, obtuvo: %v", err)
	}
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
