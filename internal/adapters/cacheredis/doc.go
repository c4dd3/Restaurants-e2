// Package cacheredis implementa el port ports.Cache contra Redis (go-redis/v9).
//
// Único tipo exportado: *Cache. Verificación al pie de cache.go:
//
//	var _ ports.Cache = (*Cache)(nil)
//
// Responsabilidades:
//   - Serializar/deserializar valores con encoding/json.
//   - Exponer Get/Set/Del/DelByPattern sobre un cliente go-redis.
//   - Convertir redis.Nil → ports.ErrCacheMiss{} (el service lo interpreta).
//
// NO responsabilidades:
//   - No decide TTL (lo elige el service caso por caso).
//   - No conoce las keys específicas ("restaurants:all", "products:cat:X")
//     — esas las arma el service.
package cacheredis
