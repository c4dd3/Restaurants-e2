// Package repomongo implementa los sub-DAOs de MongoDB.
//
// Cada tipo exportado cumple una interface de internal/ports/:
//
//	ports.UserRepository        ← UserRepoMongo
//	ports.RestaurantRepository  ← RestaurantRepoMongo
//	ports.MenuRepository        ← MenuRepoMongo
//	ports.ProductRepository     ← ProductRepoMongo
//	ports.ReservationRepository ← ReservationRepoMongo
//	ports.OrderRepository       ← OrderRepoMongo
//
// var _ ports.UserRepository = (*UserRepoMongo)(nil)   al final de cada archivo.
//
// Topología del cluster (ver deployments/docker-compose.yml):
//   - 1 config server (replica set cfgrs).
//   - 2 shards, cada uno con replica set de 3 nodos (1 primario + 2 secundarios).
//   - 1 mongos router → los clientes apuntan acá.
//
// Los sub-DAOs NO saben que está sharded — solo hablan con el mongos (:27017).
// El mongos enruta cada query al shard correcto usando la shard key.
//
// Shard keys (ver deployments/mongo/init-cluster.sh):
//   - products.category       (hashed)   → distribuye escrituras por categoría.
//   - reservations.restaurant_id (hashed) → colocación: reservas de un mismo
//                                          restaurante viven juntas → CheckAvailability
//                                          golpea UN solo shard (evita broadcast).
//
// Convenciones:
//
//  1. Cliente único *mongo.Client inyectado por constructor (ver client.go).
//  2. Cada sub-DAO se queda con *mongo.Collection específica del struct.
//  3. Serialización: tags `bson:"..."` en los structs de dominio (ya están).
//  4. IDs: usar string (no primitive.ObjectID) → coincide con Postgres (UUIDs).
//     En Mongo el `_id` es el mismo string. Opcional: mapear _id ↔ id con bson tag.
//  5. Escrituras multi-documento: usar transacciones ACID (requiere replica set,
//     que ya tenemos). Ejemplo: orders + order_items como documento embebido o
//     dos colecciones con TX.
//
// Diferencia con Postgres:
//   - Mongo naturalmente permite embeber order_items DENTRO del documento order.
//     Eso elimina la necesidad de transacción para ese caso.
//   - Evaluar en cada entidad si conviene embeber (1:N pequeño, lectura conjunta)
//     o referenciar (1:N grande, lecturas independientes).
//
// Empaquetado:
//
//	repomongo/
//	├── doc.go
//	├── client.go          ← constructor del *mongo.Client
//	├── repositories.go    ← factory NewRepositories(client, dbName) *ports.Repositories
//	├── user.go
//	├── restaurant.go
//	├── menu.go
//	├── product.go
//	├── reservation.go
//	└── order.go
package repomongo
