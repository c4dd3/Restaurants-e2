package service

// ProductService — casos de uso sobre productos individuales.
//
// Los productos se crean típicamente vía MenuService.Create (como parte de un menú),
// pero también se pueden manipular individualmente.
//
// Dependencias:
//   - ports.ProductRepository
//   - ports.Cache
//
// Métodos públicos:
//
//   GetByID(ctx, id string) (*domain.Product, error)
//     1. Cache.Get(ctx, "products:id:"+id, &p).
//     2. Si miss → ProductRepository.FindByID(ctx, id) → ErrNotFound si no existe.
//     3. Cache.Set(ctx, "products:id:"+id, p, 10*time.Minute).
//
//   ListByCategory(ctx, category string) ([]domain.Product, error)
//     1. Cache.Get(ctx, "products:cat:"+category, &list).
//     2. Si miss → ProductRepository.FindByCategory(ctx, category).
//     3. Cache.Set.
//
//   Update(ctx, userRole string, p *domain.Product) (*domain.Product, error)
//     1. Permisos: solo admin.
//     2. ProductRepository.Update(ctx, p).
//     3. Invalidar: Cache.Del("products:id:"+p.ID), Cache.DelByPattern("products:cat:*")
//        (no sabemos si cambió de categoría — mejor borrar todos los listados).
//
//   Delete(ctx, userRole, id string) error
//     1. Permisos: solo admin.
//     2. ProductRepository.Delete(ctx, id).
//     3. Invalidación igual que Update.
//
// ⚠ Consistencia con el índice de búsqueda:
//   ElasticSearch se REindexa periódicamente (POST /search/reindex).
//   No encadenamos el delete a ES desde acá para mantener los servicios
//   desacoplados. La alternativa evolutiva es un outbox pattern o un
//   evento (Kafka/NATS) al que el search-service se suscribe — fuera
//   del alcance de Etapa 2.
