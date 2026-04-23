package service

// RestaurantService — casos de uso sobre restaurantes.
//
// Dependencias:
//   - ports.RestaurantRepository
//   - ports.Cache                  (cache-aside para listados)
//
// Métodos públicos:
//
//   Create(ctx, userRole string, req CreateRestaurantRequest) (*domain.Restaurant, error)
//     1. Verificar permisos: si userRole != "admin" → ErrForbidden.
//     2. Construir domain.Restaurant con uuid + timestamps.
//     3. RestaurantRepository.Create(ctx, &r).
//     4. Cache.DelByPattern(ctx, "restaurants:*")  — invalida listados cacheados.
//     5. Devolver el restaurante creado.
//
//   GetByID(ctx, id string) (*domain.Restaurant, error)
//     1. Intentar Cache.Get(ctx, "restaurants:id:"+id, &r).
//        - si hit (nil err) → devolver.
//        - si ErrCacheMiss → seguir.
//     2. RestaurantRepository.FindByID(ctx, id).
//        - si no existe → ErrNotFound.
//     3. Cache.Set(ctx, "restaurants:id:"+id, r, 5*time.Minute).
//     4. Devolver.
//
//   List(ctx) ([]domain.Restaurant, error)
//     Mismo patrón cache-aside con key "restaurants:all".
//
// Notas:
//   - TTL típico de 5 minutos: equilibrio entre frescura y carga de BD.
//   - El chequeo de admin vive ACÁ y no solo en el middleware, así los tests
//     unitarios del service pueden validar la regla sin tocar HTTP.
//   - En Mongo, la colección NO se shardea (pocos restaurantes → colocación OK).
