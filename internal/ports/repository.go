// Package ports define las INTERFACES que la capa de servicio necesita.
//
// Regla arquitectónica clave:
//   - `service/` importa únicamente `domain/` y `ports/`.
//   - Las implementaciones concretas viven en `adapters/` y se inyectan en `cmd/*/main.go`.
//   - Cambiar entre Postgres y Mongo es cuestión de inyectar otro adapter — la lógica
//     de negocio nunca lo nota. Esto es lo que cumple el requisito "no condiciones
//     específicas por tecnología en la lógica de negocio" del enunciado.
package ports

import (
	"context"

	"restaurants-e2/internal/domain"
)

// UserRepository — operaciones sobre usuarios.
// Todos los métodos aceptan context.Context para permitir timeouts y cancelación desde el request handler.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

// RestaurantRepository — operaciones sobre restaurantes.
type RestaurantRepository interface {
	Create(ctx context.Context, r *domain.Restaurant) error
	FindByID(ctx context.Context, id string) (*domain.Restaurant, error)
	FindAll(ctx context.Context) ([]domain.Restaurant, error)
}

// MenuRepository — operaciones sobre menús.
type MenuRepository interface {
	Create(ctx context.Context, m *domain.Menu) error
	FindByID(ctx context.Context, id string) (*domain.Menu, error)
	Update(ctx context.Context, id string, req *domain.UpdateMenuRequest) (*domain.Menu, error)
	Delete(ctx context.Context, id string) error
}

// ProductRepository — operaciones sobre productos (antes MenuItem).
// En MongoDB esta colección se shardea por `category` (hashed).
type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Product, error)
	FindByCategory(ctx context.Context, category string) ([]domain.Product, error)
	FindAll(ctx context.Context) ([]domain.Product, error)
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id string) error
}

// ReservationRepository — operaciones sobre reservas.
// En MongoDB esta colección se shardea por `restaurant_id` (hashed).
type ReservationRepository interface {
	Create(ctx context.Context, r *domain.Reservation) error
	FindByID(ctx context.Context, id string) (*domain.Reservation, error)
	Cancel(ctx context.Context, id string) error
	CheckAvailability(ctx context.Context, restaurantID string, partySize int) (availableSeats int, err error)
}

// OrderRepository — operaciones sobre órdenes.
type OrderRepository interface {
	Create(ctx context.Context, o *domain.Order) error
	FindByID(ctx context.Context, id string) (*domain.Order, error)
}

// Repositories agrupa todos los repositorios. La función de wiring en main.go
// devuelve un *Repositories armado con las implementaciones elegidas.
type Repositories struct {
	Users        UserRepository
	Restaurants  RestaurantRepository
	Menus        MenuRepository
	Products     ProductRepository
	Reservations ReservationRepository
	Orders       OrderRepository
}
