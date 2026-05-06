package domain

import "time"

// Menu representa un menú perteneciente a un restaurante.
// Los productos del menú se cargan como slice de Product (antes MenuItem).
type Menu struct {
	ID           string    `json:"id"            db:"id"            bson:"_id,omitempty"`
	RestaurantID string    `json:"restaurant_id" db:"restaurant_id" bson:"restaurant_id"`
	Name         string    `json:"name"          db:"name"          bson:"name"`
	Description  string    `json:"description"   db:"description"   bson:"description"`
	Products     []Product `json:"products,omitempty" db:"-" bson:"products,omitempty"` // db:"-": no es columna SQL; en Postgres se carga por join separado.
	CreatedAt    time.Time `json:"created_at"    db:"created_at"    bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"    db:"updated_at"    bson:"updated_at"`
}
