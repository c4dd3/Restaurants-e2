package repomongo

// repositories.go — factory que construye TODOS los sub-DAOs de Mongo.
// Función pública:
//   NewRepositories(client *mongo.Client, dbName string) *ports.Repositories
// Este archivo es el espejo de repopg/repositories.go.
//  Ambos devuelven elmismo  tipo (*ports.Repositories)
// el wiring (internal/wiring) elige uno u otro según DB_ENGINE sin que el resto del sistema se entere.

import (
	"go.mongodb.org/mongo-driver/mongo"

	"restaurants-e2/internal/ports"
)

func NewRepositories(client *mongo.Client, dbName string) *ports.Repositories {
	db := client.Database(dbName)
	return &ports.Repositories{
		Users:        &UserRepoMongo{coll: db.Collection("users")},
		Restaurants:  &RestaurantRepoMongo{coll: db.Collection("restaurants")},
		Menus:        &MenuRepoMongo{db: db, coll: db.Collection("menus")},
		Products:     &ProductRepoMongo{coll: db.Collection("products")},
		Reservations: &ReservationRepoMongo{db: db, coll: db.Collection("reservations")},
		Orders:       &OrderRepoMongo{coll: db.Collection("orders")},
	}
}
