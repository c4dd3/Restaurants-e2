package domain

import "time"

// DTOs (Data Transfer Objects) para entradas y salidas de la API.
// Viven en domain/ porque describen el contrato público del negocio —
// no son específicos de HTTP ni de un motor de BD.

// RegisterRequest — POST /auth/register
// Role es opcional; si se omite o es inválido, se asigna "client" por defecto.
// Valores válidos: "client" | "admin".
type RegisterRequest struct {
	Name     string `json:"name"     binding:"required,min=2"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

// LoginRequest — POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse — respuesta de login/register exitoso.
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateUserRequest — PUT /users/:id. Campos opcionales.
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateRestaurantRequest — POST /restaurants
type CreateRestaurantRequest struct {
	Name        string `json:"name"        binding:"required"`
	Address     string `json:"address"     binding:"required"`
	Phone       string `json:"phone"       binding:"required"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"    binding:"required,min=1"`
}

// CreateMenuRequest — POST /menus
type CreateMenuRequest struct {
	RestaurantID string           `json:"restaurant_id" binding:"required"`
	Name         string           `json:"name"          binding:"required"`
	Description  string           `json:"description"`
	Products     []ProductRequest `json:"products"`
}

// ProductRequest describe un producto al crear/actualizar un menú.
type ProductRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description string  `json:"description"`
	Category    string  `json:"category"    binding:"required"`
	Price       float64 `json:"price"       binding:"required,gt=0"`
	Available   bool    `json:"available"`
}

// UpdateMenuRequest — PUT /menus/:id. Todos los campos son opcionales.
type UpdateMenuRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Products    []ProductRequest `json:"products"`
}

// CreateReservationRequest — POST /reservations
type CreateReservationRequest struct {
	RestaurantID string    `json:"restaurant_id" binding:"required"`
	Date         time.Time `json:"date"          binding:"required"`
	PartySize    int       `json:"party_size"    binding:"required,min=1"`
	Notes        string    `json:"notes"`
}

// CreateOrderRequest — POST /orders
type CreateOrderRequest struct {
	RestaurantID  string             `json:"restaurant_id" binding:"required"`
	ReservationID *string            `json:"reservation_id"`
	Items         []OrderItemRequest `json:"items"         binding:"required,min=1"`
	Pickup        bool               `json:"pickup"`
}

// OrderItemRequest — línea de pedido al crear una orden.
type OrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity"   binding:"required,min=1"`
}

// Claims es la información extraída de un JWT validado.
// El middleware de auth la guarda en el contexto de la request.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
