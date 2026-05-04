package repomongo

import (
	"context"
	"os"
	"testing"
	"time"

	"restaurants-e2/internal/config"
	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// Helper para pruebas de Mongo: usa una DB temporal y la limpia al final.
func testMongoRepositories(t *testing.T) (*ports.Repositories, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	dbName := "restaurants_unit_test"

	client, err := NewClient(ctx, config.MongoConfig{URI: uri, DBName: dbName})
	if err != nil {
		t.Skipf("MongoDB local no disponible para pruebas de repos: %v", err)
	}
	_ = client.Database(dbName).Drop(ctx)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = client.Database(dbName).Drop(ctx)
		_ = client.Disconnect(ctx)
	}
	return NewRepositories(client, dbName), cleanup
}

// CRUD básico del repo de usuarios contra Mongo real.
func TestUserRepoMongoCRUD(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	u := &domain.User{ID: "user-1", Name: "Bea", Email: "bea@example.com", Password: "hash", Role: domain.RoleClient}
	if err := repos.Users.Create(ctx, u); err != nil {
		t.Fatal(err)
	}

	byID, err := repos.Users.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if byID == nil || byID.Email != "bea@example.com" {
		t.Fatalf("usuario por id incorrecto: %#v", byID)
	}

	byEmail, err := repos.Users.FindByEmail(ctx, "bea@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if byEmail == nil || byEmail.ID != "user-1" {
		t.Fatalf("usuario por email incorrecto: %#v", byEmail)
	}

	updated, err := repos.Users.Update(ctx, "user-1", &domain.UpdateUserRequest{Name: "Beatriz"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Beatriz" {
		t.Fatalf("nombre no actualizado: %#v", updated)
	}

	if err := repos.Users.Delete(ctx, "user-1"); err != nil {
		t.Fatal(err)
	}
	missing, err := repos.Users.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Fatalf("se esperaba usuario eliminado, obtuvo %#v", missing)
	}
}

// Casos extra: ids automáticos, fechas y búsquedas que no encuentran nada.
func TestUserRepoMongoDefaultsAndMissing(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	u := &domain.User{Name: "Sin ID", Email: "sinid@example.com", Password: "hash", Role: domain.RoleAdmin}
	if err := repos.Users.Create(ctx, u); err != nil {
		t.Fatal(err)
	}
	if u.ID == "" || u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		t.Fatalf("no llenó campos automáticos: %#v", u)
	}

	missing, err := repos.Users.FindByID(ctx, "no-existe")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Fatalf("esperaba nil para usuario inexistente, obtuvo %#v", missing)
	}

	missing, err = repos.Users.FindByEmail(ctx, "nadie@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Fatalf("esperaba nil para email inexistente, obtuvo %#v", missing)
	}

	updated, err := repos.Users.Update(ctx, "no-existe", &domain.UpdateUserRequest{Name: "Nada"})
	if err != nil {
		t.Fatal(err)
	}
	if updated != nil {
		t.Fatalf("esperaba nil al actualizar usuario inexistente, obtuvo %#v", updated)
	}

	// Borrar un id inexistente no debería romper el flujo.
	if err := repos.Users.Delete(ctx, "no-existe"); err != nil {
		t.Fatal(err)
	}
}

func TestMongoClientError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := NewClient(ctx, config.MongoConfig{URI: "mongodb://127.0.0.1:1", DBName: "restaurants_unit_test"})
	if err == nil {
		t.Fatal("esperaba error conectando a un Mongo inexistente")
	}
}
