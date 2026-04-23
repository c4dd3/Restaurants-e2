package repopg

// product.go — sub-DAO de productos para Postgres.
//
// Struct:  type ProductRepoPg struct { pool *pgxpool.Pool }
// Verify:  var _ ports.ProductRepository = (*ProductRepoPg)(nil)
//
// Tabla: products
// Columnas: id, menu_id (FK), name, description, category, price, available,
//           created_at, updated_at.
// Índice: CREATE INDEX idx_products_category ON products(category)  (ver init.sql).
//
// Métodos:
//
// FindByID(ctx, id) (*domain.Product, error)
//   SELECT * FROM products WHERE id = $1;
//
// FindByCategory(ctx, category) ([]domain.Product, error)
//   SELECT * FROM products WHERE category = $1 ORDER BY name;
//   (Usa el índice idx_products_category.)
//
// FindAll(ctx) ([]domain.Product, error)
//   SELECT * FROM products ORDER BY created_at DESC;
//   Si el catálogo crece, agregar paginación (LIMIT/OFFSET → keyset pagination).
//
// Create(ctx, p) error
//   INSERT INTO products (id, menu_id, name, description, category, price, available)
//   VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at;
//   Si p.Description == "" → dejarlo vacío; el getter DescriptionOrDefault()
//   del dominio aplica "Producto sin descripción" al leer. NO escribir el
//   default en BD (mantiene el dato puro).
//
// Update(ctx, p) error
//   UPDATE products SET name=$2, description=$3, category=$4, price=$5,
//     available=$6, updated_at=now() WHERE id=$1;
//
// Delete(ctx, id) error
//   DELETE FROM products WHERE id = $1;
//
// Observación para la defensa:
//   En Mongo esta misma colección se shardea por `category` (hashed). Eso
//   distribuye bien las escrituras (muchas categorías diversas) sin romper
//   las queries por categoría (el mongos enruta al shard correcto).
