package repopg

// repositories.go — factory que devuelve un *ports.Repositories armado con
// TODOS los sub-DAOs de Postgres.
//
// Función pública:
//
//   NewRepositories(pool *pgxpool.Pool) *ports.Repositories
//
// Lógica:
//   Construir cada sub-DAO con el mismo pool y devolverlo en el struct:
//
//     return &ports.Repositories{
//         Users:        &UserRepoPg{pool: pool},
//         Restaurants:  &RestaurantRepoPg{pool: pool},
//         Menus:        &MenuRepoPg{pool: pool},
//         Products:     &ProductRepoPg{pool: pool},
//         Reservations: &ReservationRepoPg{pool: pool},
//         Orders:       &OrderRepoPg{pool: pool},
//     }
//
// Este archivo es el "padre pegamento" de Postgres: centraliza el wiring
// de todos los sub-DAOs. Lo consume internal/wiring/repositories.go.
//
// Por qué NO mezclar lógica acá:
//   Este archivo es una pura factoría — sin queries, sin lógica. Si crece,
//   significa que alguien está metiendo código que no corresponde.
