package cacheredis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"restaurants-e2/internal/ports"
)

var _ ports.Cache = (*Cache)(nil)

// Cache implementa ports.Cache usando Redis como backend.
// Serializa y deserializa valores con JSON, lo que permite cachear cualquier
// struct de dominio sin cambios en el adapter.
type Cache struct {
	rdb *redis.Client
}

// New construye un Cache a partir de un cliente ya inicializado.
func New(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

// Get deserializa en dest el valor almacenado bajo key.
// Retorna ports.ErrCacheMiss si la clave no existe (cache miss esperado).
func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ports.ErrCacheMiss{}
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}

// Set serializa value a JSON y lo almacena con el TTL dado.
// Si ttl es 0, Redis usa la política de expiración configurada (allkeys-lru).
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

// Del elimina una o más claves.
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// DelByPattern elimina todas las claves que coincidan con el patrón.
// Usa SCAN en lotes (nunca KEYS) para no bloquear Redis en producción.
func (c *Cache) DelByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var toDelete []string

	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, pattern, 500).Result()
		if err != nil {
			return err
		}
		toDelete = append(toDelete, keys...)
		if next == 0 {
			break
		}
		cursor = next
	}

	if len(toDelete) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, toDelete...).Err()
}
