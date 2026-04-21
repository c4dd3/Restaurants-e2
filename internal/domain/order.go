package domain

import "time"

// Order representa un pedido de comida realizado por un usuario.
// ReservationID es opcional (puntero a string): puede haber pedidos sin reserva previa (takeaway).
type Order struct {
	ID            string      `json:"id"                        db:"id"            bson:"_id,omitempty"`
	UserID        string      `json:"user_id"                   db:"user_id"       bson:"user_id"`
	RestaurantID  string      `json:"restaurant_id"             db:"restaurant_id" bson:"restaurant_id"`
	ReservationID *string     `json:"reservation_id,omitempty"  db:"reservation_id" bson:"reservation_id,omitempty"`
	Items         []OrderItem `json:"items,omitempty"           bson:"items,omitempty"`
	Total         float64     `json:"total"                     db:"total"         bson:"total"`
	Status        string      `json:"status"                    db:"status"        bson:"status"`
	Pickup        bool        `json:"pickup"                    db:"pickup"        bson:"pickup"`
	CreatedAt     time.Time   `json:"created_at"                db:"created_at"    bson:"created_at"`
}

// OrderItem es una línea del pedido. Price se guarda al momento de crear el pedido:
// si el producto cambia de precio después, el pedido original conserva el original.
type OrderItem struct {
	ID        string  `json:"id"         db:"id"         bson:"_id,omitempty"`
	OrderID   string  `json:"order_id"   db:"order_id"   bson:"order_id"`
	ProductID string  `json:"product_id" db:"product_id" bson:"product_id"`
	Quantity  int     `json:"quantity"   db:"quantity"   bson:"quantity"`
	Price     float64 `json:"price"      db:"price"      bson:"price"`
}
