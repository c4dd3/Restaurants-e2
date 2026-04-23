package repopg

// menu.go — sub-DAO de menús para Postgres.
//
// Struct:  type MenuRepoPg struct { pool *pgxpool.Pool }
// Verify:  var _ ports.MenuRepository = (*MenuRepoPg)(nil)
//
// Tabla: menus
// Columnas: id, restaurant_id (FK), name, description, created_at, updated_at.
//
// Métodos:
//
// Create(ctx, m) error
//   INSERT INTO menus (id, restaurant_id, name, description)
//   VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at;
//   NOTA: este método solo crea el menú. Los productos asociados se crean
//   por separado con ProductRepoPg.Create. Si se quiere hacer atómico menú+productos,
//   el service debe exponer un método más alto (CreateMenuWithProducts) que
//   envuelva todo en pool.BeginTx.
//
// FindByID(ctx, id) (*domain.Menu, error)
//   Para devolver el menú CON sus productos:
//     1. SELECT * FROM menus WHERE id = $1.
//     2. SELECT * FROM products WHERE menu_id = $1 ORDER BY name.
//     3. Asignar m.Products = productsSlice.
//   Alternativa con 1 query: LEFT JOIN + pgx.CollectRows con agrupación manual
//   (más rápido, más código — evaluar con benchmarks).
//
// Update(ctx, id, req) (*domain.Menu, error)
//   Patrón COALESCE como en user.go.
//   Si req.Products != nil → reemplazar productos:
//     - DELETE FROM products WHERE menu_id = $id;
//     - INSERT de cada producto nuevo.
//     - Todo dentro de una transacción.
//
// Delete(ctx, id) error
//   El schema tiene ON DELETE CASCADE sobre products.menu_id → borrar menú
//   borra productos automáticamente. Solo:
//     DELETE FROM menus WHERE id = $1;
