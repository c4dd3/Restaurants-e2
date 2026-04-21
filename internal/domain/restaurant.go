package domain

import "time"

// Restaurant representa un establecimiento registrado en el sistema.
// AdminID referencia al User dueño del restaurante.
type Restaurant struct {
	ID          string    `json:"id"          db:"id"          bson:"_id,omitempty"`
	Name        string    `json:"name"        db:"name"        bson:"name"`
	Address     string    `json:"address"     db:"address"     bson:"address"`
	Phone       string    `json:"phone"       db:"phone"       bson:"phone"`
	Description string    `json:"description" db:"description" bson:"description"`
	AdminID     string    `json:"admin_id"    db:"admin_id"    bson:"admin_id"`
	Capacity    int       `json:"capacity"    db:"capacity"    bson:"capacity"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"  bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"  bson:"updated_at"`
}
