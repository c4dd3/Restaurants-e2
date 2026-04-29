package repopg

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"restaurants-e2/internal/ports"
)

// NewRepositories arma los 6 sub-DAOs con el mismo pool y los devuelve
// en el struct ports.Repositories que consume el service layer.
// Es una pura factoría — sin queries ni lógica.
func NewRepositories(pool *pgxpool.Pool) *ports.Repositories {
	return &ports.Repositories{
		Users:        &UserRepoPg{pool: pool},
		Restaurants:  &RestaurantRepoPg{pool: pool},
		Menus:        &MenuRepoPg{pool: pool},
		Products:     &ProductRepoPg{pool: pool},
		Reservations: &ReservationRepoPg{pool: pool},
		Orders:       &OrderRepoPg{pool: pool},
	}
}
