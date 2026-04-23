package service

// MenuService — casos de uso sobre menús.
//
// Un menú pertenece a un restaurante y contiene una lista de productos.
//
// Dependencias:
//   - ports.MenuRepository
//   - ports.RestaurantRepository   (para validar que el restaurante exista)
//   - ports.ProductRepository      (para persistir los productos del menú)
//   - ports.Cache
//
// Métodos públicos:
//
//   Create(ctx, userRole string, req CreateMenuRequest) (*domain.Menu, error)
//     1. Chequear permisos: si userRole != "admin" → ErrForbidden.
//     2. Validar que el restaurante exista (RestaurantRepository.FindByID).
//        Si no → ErrValidation (referencia inválida).
//     3. Construir domain.Menu (uuid + restaurant_id + name + desc).
//     4. MenuRepository.Create(ctx, &m).
//     5. Para cada ProductRequest en req.Products:
//        - Si Description vacío → aplicar "Producto sin descripción"
//          (o dejarlo y que el getter DescriptionOrDefault() decida).
//        - Construir domain.Product con MenuID = m.ID.
//        - ProductRepository.Create(ctx, &p).
//     6. Cache.DelByPattern(ctx, "products:*"). Los listados por categoría cambian.
//     7. Devolver el menú con sus productos.
//
//   GetByID / Update / Delete: CRUD estándar, mismo patrón cache-aside.
//
// ⚠ Consistencia multi-documento:
//   En Postgres → transacción BEGIN/COMMIT cubre Menu + Products.
//   En Mongo    → multi-document transaction (requiere replica set, que ya tenemos).
//   Ese detalle transaccional vive en el adapter, no acá. El service solo ve
//   "si falla a medias, devolver error". El repo es quien envuelve la TX.
