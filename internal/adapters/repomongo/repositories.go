package repomongo

// repositories.go — factory que construye TODOS los sub-DAOs de Mongo.
//
// Función pública:
//
//   NewRepositories(client *mongo.Client, dbName string) *ports.Repositories
//
// Lógica:
//
//   db := client.Database(dbName)  // típicamente "restaurants"
//
//   return &ports.Repositories{
//       Users:        &UserRepoMongo{coll: db.Collection("users")},
//       Restaurants:  &RestaurantRepoMongo{coll: db.Collection("restaurants")},
//       Menus:        &MenuRepoMongo{coll: db.Collection("menus")},
//       Products:     &ProductRepoMongo{coll: db.Collection("products")},
//       Reservations: &ReservationRepoMongo{coll: db.Collection("reservations")},
//       Orders:       &OrderRepoMongo{coll: db.Collection("orders")},
//   }
//
// Este archivo es el espejo de repopg/repositories.go. Ambos devuelven el
// MISMO tipo (*ports.Repositories) — el wiring (internal/wiring) elige uno
// u otro según DB_ENGINE sin que el resto del sistema se entere.
