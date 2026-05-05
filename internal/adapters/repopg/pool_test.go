package repopg

import (
	"context"
	"os"
	"testing"
	"time"

	"restaurants-e2/internal/config"
)

// TestNewPool verifica que NewPool conecta correctamente a Postgres.
// Se salta si Postgres no está disponible (igual que testPool).
func TestNewPool(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/restaurants"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, config.PostgresConfig{DSN: dsn})
	if err != nil {
		t.Skipf("Postgres no disponible para TestNewPool: %v", err)
	}
	t.Cleanup(pool.Close)

	if pool == nil {
		t.Fatal("NewPool retornó nil sin error")
	}
}

// TestNewPoolBadDSN verifica que NewPool retorna error cuando Postgres no responde.
// Usa 127.0.0.1:1 (puerto reservado, sin proceso) para evitar DNS y fallar rápido.
func TestNewPoolBadDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := NewPool(ctx, config.PostgresConfig{DSN: "postgres://postgres:postgres@127.0.0.1:1/restaurants"})
	if err == nil {
		t.Fatal("NewPool debería haber fallado con puerto inaccesible")
	}
}

// TestNewRepositories verifica que NewRepositories construye el struct con todos los repos.
func TestNewRepositories(t *testing.T) {
	pool := testPool(t) // se salta si Postgres no está disponible

	repos := NewRepositories(pool)
	if repos == nil {
		t.Fatal("NewRepositories retornó nil")
	}
	if repos.Users == nil {
		t.Error("repos.Users es nil")
	}
	if repos.Restaurants == nil {
		t.Error("repos.Restaurants es nil")
	}
	if repos.Menus == nil {
		t.Error("repos.Menus es nil")
	}
	if repos.Products == nil {
		t.Error("repos.Products es nil")
	}
	if repos.Reservations == nil {
		t.Error("repos.Reservations es nil")
	}
	if repos.Orders == nil {
		t.Error("repos.Orders es nil")
	}
}
