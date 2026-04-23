package cacheredis

// cache.go — implementación de ports.Cache.
//
// Struct:
//   type Cache struct { rdb *redis.Client }
// Verify al pie:
//   var _ ports.Cache = (*Cache)(nil)
//
// Constructor:
//   func New(rdb *redis.Client) *Cache { return &Cache{rdb: rdb} }
//
// Métodos a implementar:
//
// Get(ctx, key, dest) error
//   raw, err := c.rdb.Get(ctx, key).Bytes()
//   if errors.Is(err, redis.Nil) → return ports.ErrCacheMiss{}
//   if err != nil → return err
//   return json.Unmarshal(raw, dest)
//   ⚠ `dest` debe ser un puntero (ej: &userVar, &[]domain.User).
//
// Set(ctx, key, value, ttl) error
//   data, err := json.Marshal(value)
//   if err != nil → return err
//   return c.rdb.Set(ctx, key, data, ttl).Err()
//
// Del(ctx, keys...) error
//   return c.rdb.Del(ctx, keys...).Err()
//
// DelByPattern(ctx, pattern) error
//   Usar SCAN en lote (NO el comando KEYS — bloqueante en BD grandes).
//
//     var cursor uint64
//     var batch []string
//     for {
//         keys, next, err := c.rdb.Scan(ctx, cursor, pattern, 500).Result()
//         if err != nil { return err }
//         batch = append(batch, keys...)
//         if next == 0 { break }
//         cursor = next
//     }
//     if len(batch) == 0 { return nil }
//     return c.rdb.Del(ctx, batch...).Err()
//
//   Para optimizar si el patrón tiene MUCHISIMAS claves, se puede hacer
//   Del incremental dentro del loop (de a 500). Para Etapa 2 alcanza esto.
//
// Notas:
//   - Redis tiene maxmemory-policy allkeys-lru (ver docker-compose.yml) →
//     al llenarse 256MB, evicta las claves usadas menos recientemente.
//     Eso significa que NO hay que preocuparse por "llenar el caché" — se
//     auto-podará.
//   - No guardar datos sensibles (passwords, tokens). Solo lecturas derivadas.
