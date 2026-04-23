// Package repopg implementa los sub-DAOs de PostgreSQL.
//
// Cada tipo exportado (UserRepoPg, RestaurantRepoPg, ...) es un SUB-DAO que
// cumple una interface definida en internal/ports/ (el "DAO padre"):
//
//	ports.UserRepository        ← UserRepoPg
//	ports.RestaurantRepository  ← RestaurantRepoPg
//	ports.MenuRepository        ← MenuRepoPg
//	ports.ProductRepository     ← ProductRepoPg
//	ports.ReservationRepository ← ReservationRepoPg
//	ports.OrderRepository       ← OrderRepoPg
//
// Verificación en tiempo de compilación:
//
//	var _ ports.UserRepository = (*UserRepoPg)(nil)
//
// Si un método falta o tiene firma equivocada, el proyecto no compila.
// Esa línea va al final de cada archivo de sub-DAO.
//
// Convenciones:
//
//  1. Todos los sub-DAOs comparten un *pgxpool.Pool — se construye una sola vez
//     en pool.go y se inyecta por constructor.
//  2. Nunca concatenar strings para armar SQL — siempre parámetros posicionales
//     ($1, $2, ...). pgx los pasa al server con el protocolo extended (previene
//     inyección SQL y ayuda al plan cache).
//  3. Escaneo con pgx.RowToStructByName[T] — mapea columnas con tags `db:"..."`
//     de los structs del dominio. Sin librerías ORM.
//  4. Para inserciones múltiples relacionadas (p.ej. order + order_items) usar
//     pool.BeginTx + defer tx.Rollback + tx.Commit al final.
//  5. Errores traducidos a pgconn.PgError; 23505 = unique violation → propagar
//     como error genérico, el service lo mapeará a ErrConflict.
//
// Empaquetado:
//
//	repopg/
//	├── doc.go            ← esto
//	├── pool.go           ← constructor del *pgxpool.Pool
//	├── repositories.go   ← factory NewRepositories(pool) *ports.Repositories
//	├── user.go           ← sub-DAO
//	├── restaurant.go     ← sub-DAO
//	├── menu.go           ← sub-DAO
//	├── product.go        ← sub-DAO
//	├── reservation.go    ← sub-DAO
//	└── order.go          ← sub-DAO
package repopg
