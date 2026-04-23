package repopg

// user.go — sub-DAO de usuarios para Postgres.
//
// Struct:
//
//   type UserRepoPg struct { pool *pgxpool.Pool }
//
// Implementa ports.UserRepository. Al final del archivo:
//
//   var _ ports.UserRepository = (*UserRepoPg)(nil)
//
// Tabla: users (ver deployments/postgres/init.sql).
// Columnas: id (uuid), name, email (unique), password_hash, role, created_at, updated_at.
//
// Métodos a implementar:
//
// FindByID(ctx, id) (*domain.User, error)
//   SQL: SELECT id, name, email, password_hash, role, created_at, updated_at
//        FROM users WHERE id = $1
//   Scan con pgx.RowToStructByName[domain.User].
//   Si pgx.ErrNoRows → devolver nil, nil (el service lo convierte a ErrNotFound).
//
// FindByEmail(ctx, email) (*domain.User, error)
//   SQL: SELECT ... FROM users WHERE email = $1
//   Mismo tratamiento. Devuelve nil, nil si no existe.
//
// Create(ctx, u) error
//   SQL: INSERT INTO users (id, name, email, password_hash, role)
//        VALUES ($1, $2, $3, $4, $5)
//        RETURNING created_at, updated_at
//   Si u.ID está vacío, generarlo con uuid.New().String() ANTES del INSERT.
//   Si pgconn.PgError con SQLState == "23505" (unique_violation) → devolver
//   error envuelto; el service lo mapea a ErrConflict.
//
// Update(ctx, id, req) (*domain.User, error)
//   Construir SET dinámico solo con los campos no vacíos del DTO.
//   Patrón seguro: usar COALESCE:
//     UPDATE users
//        SET name  = COALESCE(NULLIF($2, ''), name),
//            email = COALESCE(NULLIF($3, ''), email),
//            updated_at = now()
//      WHERE id = $1
//      RETURNING *
//
// Delete(ctx, id) error
//   SQL: DELETE FROM users WHERE id = $1
//   Usar pool.Exec y chequear RowsAffected — si == 0 → nil (idempotente) o
//   ErrNotFound (depende de cómo el service lo quiera interpretar).
