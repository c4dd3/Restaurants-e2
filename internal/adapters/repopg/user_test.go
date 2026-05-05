package repopg

// user_test.go — pruebas de integración contra Postgres real.
// Se saltan automáticamente si POSTGRES_TEST_URL no está disponible.
// Para correr localmente con el stack levantado:
//   POSTGRES_TEST_URL="postgres://postgres:postgres@localhost:5432/restaurants" go test ./internal/adapters/repopg/...

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"restaurants-e2/internal/domain"
)

// testPool conecta a Postgres usando POSTGRES_TEST_URL.
// Si la variable no está, salta el test con t.Skip.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("POSTGRES_TEST_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/restaurants"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Skipf("DSN inválido, skipping integración Postgres: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Skipf("no se pudo conectar a Postgres, skipping: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Postgres no disponible, skipping: %v", err)
	}

	t.Cleanup(pool.Close)
	return pool
}

// TestUserRepoPgCRUD cubre el ciclo completo: Create → FindByID → FindByEmail → Update → Delete.
func TestUserRepoPgCRUD(t *testing.T) {
	pool := testPool(t)
	repo := NewUserRepoPg(pool)
	ctx := context.Background()

	u := &domain.User{
		ID:       uuid.NewString(),
		Name:     "Catalina Brenes",
		Email:    "catalina-pg-" + uuid.NewString()[:8] + "@example.com",
		Password: "hashed-password",
		Role:     domain.RoleClient,
	}

	// t.Cleanup garantiza limpieza aunque el test falle a mitad
	t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })

	// Create
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		t.Error("Create no llenó timestamps")
	}

	// FindByID
	found, err := repo.FindByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Email != u.Email {
		t.Errorf("email esperado %q, obtenido %q", u.Email, found.Email)
	}

	// FindByEmail
	byEmail, err := repo.FindByEmail(ctx, u.Email)
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if byEmail.ID != u.ID {
		t.Errorf("ID esperado %q, obtenido %q", u.ID, byEmail.ID)
	}

	// Update
	updated, err := repo.Update(ctx, u.ID, &domain.UpdateUserRequest{Name: "Catalina Brenes Mora"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Catalina Brenes Mora" {
		t.Errorf("nombre no actualizado: %q", updated.Name)
	}

	// Delete
	if err := repo.Delete(ctx, u.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Verificar que ya no existe
	_, err = repo.FindByID(ctx, u.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound después de Delete, obtenido %v", err)
	}
}

// TestUserRepoPgDuplicateEmail verifica que insertar el mismo email retorna ErrConflict.
func TestUserRepoPgDuplicateEmail(t *testing.T) {
	pool := testPool(t)
	repo := NewUserRepoPg(pool)
	ctx := context.Background()

	sharedEmail := "dup-pg-" + uuid.NewString()[:8] + "@example.com"
	u1 := &domain.User{ID: uuid.NewString(), Name: "User A", Email: sharedEmail, Password: "h", Role: domain.RoleClient}
	u2 := &domain.User{ID: uuid.NewString(), Name: "User B", Email: sharedEmail, Password: "h", Role: domain.RoleClient}

	if err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("primer Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, u1.ID) })

	err := repo.Create(ctx, u2)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("esperado ErrConflict por email duplicado, obtenido %v", err)
	}
}

// TestUserRepoPgFindMissing verifica que buscar un ID inexistente retorna ErrNotFound.
func TestUserRepoPgFindMissing(t *testing.T) {
	pool := testPool(t)
	repo := NewUserRepoPg(pool)
	ctx := context.Background()

	// UUIDs válidos pero que no existen en la BD
	_, err := repo.FindByID(ctx, uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("FindByID: esperado ErrNotFound, obtenido %v", err)
	}

	_, err = repo.FindByEmail(ctx, "nadie-"+uuid.NewString()+"@nowhere.com")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("FindByEmail: esperado ErrNotFound, obtenido %v", err)
	}
}

// TestUserRepoPgDeleteNotFound verifica que borrar un ID inexistente retorna ErrNotFound.
func TestUserRepoPgDeleteNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewUserRepoPg(pool)
	ctx := context.Background()

	// UUID válido que no existe
	err := repo.Delete(ctx, uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete: esperado ErrNotFound, obtenido %v", err)
	}
}
