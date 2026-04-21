package domain

import "time"

// Estados compartidos entre Reservation y Order.
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
)

// Reservation representa una reserva de mesa.
//
// Nota sobre sharding en MongoDB:
// La shard key recomendada es `restaurant_id` (hashed) — las consultas más comunes
// filtran por restaurante (ej: disponibilidad), así que mantener todas las reservas de
// un restaurante juntas evita scatter-gather.
type Reservation struct {
	ID           string    `json:"id"            db:"id"            bson:"_id,omitempty"`
	RestaurantID string    `json:"restaurant_id" db:"restaurant_id" bson:"restaurant_id"`
	UserID       string    `json:"user_id"       db:"user_id"       bson:"user_id"`
	Date         time.Time `json:"date"          db:"date"          bson:"date"`
	PartySize    int       `json:"party_size"    db:"party_size"    bson:"party_size"`
	Status       string    `json:"status"        db:"status"        bson:"status"`
	Notes        string    `json:"notes"         db:"notes"         bson:"notes"`
	CreatedAt    time.Time `json:"created_at"    db:"created_at"    bson:"created_at"`
}
