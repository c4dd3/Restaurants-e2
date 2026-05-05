package cacheredis

import (
	"context"
	"errors"
	"testing"
	"time"

	"restaurants-e2/internal/config"
	"restaurants-e2/internal/ports"
)

// testRedisClient crea un cliente Redis real para pruebas.
// Salta el test si Redis no está disponible en el entorno.
func testRedisClient(t *testing.T) *Cache {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := NewClient(ctx, config.RedisConfig{
		Addr: "localhost:6379",
	})
	if err != nil {
		t.Skipf("Redis no disponible para pruebas: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return New(client)
}

func TestCacheGetMiss(t *testing.T) {
	c := testRedisClient(t)
	ctx := context.Background()

	var dest string
	err := c.Get(ctx, "clave-inexistente-xyz", &dest)
	if !errors.As(err, &ports.ErrCacheMiss{}) {
		t.Fatalf("Get en cache vacío debió retornar ErrCacheMiss, obtuvo %v", err)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	c := testRedisClient(t)
	ctx := context.Background()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	key := "test:cache:setget"
	defer c.Del(ctx, key)

	if err := c.Set(ctx, key, payload{Name: "Bea", Age: 22}, 5*time.Second); err != nil {
		t.Fatalf("Set falló: %v", err)
	}

	var out payload
	if err := c.Get(ctx, key, &out); err != nil {
		t.Fatalf("Get tras Set falló: %v", err)
	}
	if out.Name != "Bea" || out.Age != 22 {
		t.Fatalf("valor recuperado incorrecto: %+v", out)
	}
}

func TestCacheDel(t *testing.T) {
	c := testRedisClient(t)
	ctx := context.Background()

	key := "test:cache:del"
	if err := c.Set(ctx, key, "valor", 5*time.Second); err != nil {
		t.Fatalf("Set falló: %v", err)
	}

	if err := c.Del(ctx, key); err != nil {
		t.Fatalf("Del falló: %v", err)
	}

	var dest string
	err := c.Get(ctx, key, &dest)
	if !errors.As(err, &ports.ErrCacheMiss{}) {
		t.Fatalf("tras Del esperaba ErrCacheMiss, obtuvo %v", err)
	}
}

func TestCacheDelEmpty(t *testing.T) {
	c := testRedisClient(t)
	ctx := context.Background()

	// Del sin claves no debe fallar (rama vacía).
	if err := c.Del(ctx); err != nil {
		t.Fatalf("Del sin claves falló: %v", err)
	}
}

func TestCacheDelByPattern(t *testing.T) {
	c := testRedisClient(t)
	ctx := context.Background()

	prefix := "test:pattern:del:"
	keys := []string{prefix + "a", prefix + "b", prefix + "c"}
	for _, k := range keys {
		if err := c.Set(ctx, k, "x", 5*time.Second); err != nil {
			t.Fatalf("Set %s falló: %v", k, err)
		}
	}

	if err := c.DelByPattern(ctx, prefix+"*"); err != nil {
		t.Fatalf("DelByPattern falló: %v", err)
	}

	var dest string
	for _, k := range keys {
		err := c.Get(ctx, k, &dest)
		if !errors.As(err, &ports.ErrCacheMiss{}) {
			t.Fatalf("tras DelByPattern esperaba miss en %s, obtuvo %v", k, err)
		}
	}
}

func TestCacheDelByPatternNoMatches(t *testing.T) {
	c := testRedisClient(t)
	ctx := context.Background()

	// Patrón que no coincide con nada — rama len(toDelete)==0.
	if err := c.DelByPattern(ctx, "test:pattern:vacio:xyz:*"); err != nil {
		t.Fatalf("DelByPattern sin matches falló: %v", err)
	}
}

func TestNewClientBadAddr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := NewClient(ctx, config.RedisConfig{Addr: "127.0.0.1:1"})
	if err == nil {
		t.Fatal("NewClient con addr inválida debió retornar error")
	}
}
