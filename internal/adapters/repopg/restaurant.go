package repopg

// restaurant.go — sub-DAO de restaurantes para Postgres.
//
// Struct:  type RestaurantRepoPg struct { pool *pgxpool.Pool }
// Verify:  var _ ports.RestaurantRepository = (*RestaurantRepoPg)(nil)
//
// Tabla: restaurants
// Columnas: id, name, address, phone, description, capacity, created_at, updated_at.
//
// Métodos:
//
// Create(ctx, r) error
//   INSERT INTO restaurants (id, name, address, phone, description, capacity)
//   VALUES ($1, $2, $3, $4, $5, $6)
//   RETURNING created_at, updated_at;
//   Generar uuid si r.ID está vacío.
//
// FindByID(ctx, id) (*domain.Restaurant, error)
//   SELECT * FROM restaurants WHERE id = $1;
//   pgx.ErrNoRows → devolver nil, nil.
//
// FindAll(ctx) ([]domain.Restaurant, error)
//   SELECT * FROM restaurants ORDER BY created_at DESC;
//   Collect con pgx.CollectRows(rows, pgx.RowToStructByName[domain.Restaurant]).
//   Si no hay filas → devolver slice vacío, NO nil (evita nil checks upstream).
//
// Consideraciones:
//   - No hay caché acá: la cache-aside vive en el service, no en el repo.
//     El sub-DAO es "tonto": recibe query, ejecuta, mapea, devuelve.
//   - Si en el futuro se agrega paginación, la firma de FindAll cambia al
//     port primero, y luego todos los sub-DAOs (pg y mongo) la implementan.
