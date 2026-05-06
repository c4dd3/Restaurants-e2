package cacheredis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"restaurants-e2/internal/config"
	"restaurants-e2/internal/ports"
)

// startMiniRedis levanta un servidor Redis en memoria para pruebas.
// No requiere Docker ni infraestructura externa.
func startMiniRedis(t *testing.T) *Cache {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("no se pudo iniciar miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client, err := NewClient(context.Background(), config.RedisConfig{Addr: mr.Addr()})
	if err != nil {
		t.Fatalf("no se pudo conectar a miniredis: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return New(client)
}

func TestCacheGetMiss(t *testing.T) {
	c := startMiniRedis(t)
	var dest string
	err := c.Get(context.Background(), "clave-inexistente", &dest)
	if !errors.As(err, &ports.ErrCacheMiss{}) {
		t.Fatalf("esperaba ErrCacheMiss en cache vacío, obtuvo %v", err)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	c := startMiniRedis(t)
	ctx := context.Background()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := c.Set(ctx, "test:setget", payload{Name: "Bea", Age: 22}, 5*time.Second); err != nil {
		t.Fatalf("Set falló: %v", err)
	}

	var out payload
	if err := c.Get(ctx, "test:setget", &out); err != nil {
		t.Fatalf("Get tras Set falló: %v", err)
	}
	if out.Name != "Bea" || out.Age != 22 {
		t.Fatalf("valor recuperado incorrecto: %+v", out)
	}
}

func TestCacheDel(t *testing.T) {
	c := startMiniRedis(t)
	ctx := context.Background()

	if err := c.Set(ctx, "test:del", "valor", 5*time.Second); err != nil {
		t.Fatalf("Set falló: %v", err)
	}
	if err := c.Del(ctx, "test:del"); err != nil {
		t.Fatalf("Del falló: %v", err)
	}

	var dest string
	err := c.Get(ctx, "test:del", &dest)
	if !errors.As(err, &ports.ErrCacheMiss{}) {
		t.Fatalf("tras Del esperaba ErrCacheMiss, obtuvo %v", err)
	}
}

func TestCacheDelNoKeys(t *testing.T) {
	c := startMiniRedis(t)
	// Del sin claves cubre la rama de guarda len(keys)==0.
	if err := c.Del(context.Background()); err != nil {
		t.Fatalf("Del sin claves falló: %v", err)
	}
}

func TestCacheDelByPattern(t *testing.T) {
	c := startMiniRedis(t)
	ctx := context.Background()

	prefix := "test:pattern:"
	for _, k := range []string{prefix + "a", prefix + "b", prefix + "c"} {
		if err := c.Set(ctx, k, "x", 5*time.Second); err != nil {
			t.Fatalf("Set %s falló: %v", k, err)
		}
	}

	if err := c.DelByPattern(ctx, prefix+"*"); err != nil {
		t.Fatalf("DelByPattern falló: %v", err)
	}

	var dest string
	for _, k := range []string{prefix + "a", prefix + "b", prefix + "c"} {
		err := c.Get(ctx, k, &dest)
		if !errors.As(err, &ports.ErrCacheMiss{}) {
			t.Fatalf("tras DelByPattern esperaba miss en %s, obtuvo %v", k, err)
		}
	}
}

func TestCacheDelByPatternNoMatches(t *testing.T) {
	c := startMiniRedis(t)
	// Cubre la rama len(toDelete)==0 cuando el patrón no coincide con nada.
	if err := c.DelByPattern(context.Background(), "patron:que:no:existe:*"); err != nil {
		t.Fatalf("DelByPattern sin matches falló: %v", err)
	}
}

func TestNewClientBadAddr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := NewClient(ctx, config.RedisConfig{Addr: "127.0.0.1:1"})
	if err == nil {
		t.Fatal("NewClient con dirección inválida debió retornar error")
	}
}

// TestCacheGetUnmarshalError cubre la rama json.Unmarshal en Get:
// cuando la clave existe pero el valor almacenado no es JSON válido para el tipo dest.
func TestCacheGetUnmarshalError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("no se pudo iniciar miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	// Inyectamos JSON que no puede deserializarse en un struct tipado.
	mr.Set("bad:json", "{json roto: sin comillas}")

	client, err := NewClient(context.Background(), config.RedisConfig{Addr: mr.Addr()})
	if err != nil {
		t.Fatalf("no se pudo conectar: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	cache := New(client)

	var dest struct{ X int }
	err = cache.Get(context.Background(), "bad:json", &dest)
	if err == nil {
		t.Fatal("esperaba error de unmarshal, obtuvo nil")
	}
	if errors.As(err, &ports.ErrCacheMiss{}) {
		t.Fatal("error de unmarshal no debe ser ErrCacheMiss")
	}
}
