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
