package ports

import (
	"context"
	"time"
)

// Cache abstrae el caché (Redis en la implementación actual).
// Se define aquí para que los servicios puedan envolverse con un decorador
// de caché sin depender directamente de go-redis.
//
// Uso típico:
//   1. El service intenta obtener del caché con Get(ctx, key, &dest).
//   2. Si falla (cache miss), consulta al repo, y guarda con Set(ctx, key, value, ttl).
//   3. En escrituras, invalida con Del(ctx, key).
type Cache interface {
	// Get deserializa el valor asociado a `key` en `dest` (típicamente un puntero a struct).
	// Devuelve ErrCacheMiss si la clave no existe, o un error real si falla la comunicación.
	Get(ctx context.Context, key string, dest any) error

	// Set serializa `value` y lo guarda con un TTL. Si ttl es 0 se usa el default configurado.
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Del borra una o más claves.
	Del(ctx context.Context, keys ...string) error

	// DelByPattern borra todas las claves que matcheen un patrón (ej: "products:*").
	// Útil para invalidar colecciones enteras al crear/actualizar recursos.
	DelByPattern(ctx context.Context, pattern string) error
}

// ErrCacheMiss se retorna cuando una clave no existe en caché.
// Es un error esperado, no un fallo — el caller lo usa para decidir si va al repo.
type ErrCacheMiss struct{}

func (ErrCacheMiss) Error() string { return "cache miss" }
