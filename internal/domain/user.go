// Package domain contiene las entidades puras del negocio.
// No tiene dependencias hacia adaptadores, frameworks web ni drivers de BD.
// Esto es clave para que la lógica de negocio sea agnóstica del motor (Postgres/Mongo).
package domain

import "time"

// Roles soportados por el sistema.
const (
	RoleClient = "client" // Usuario regular, puede reservar y ordenar.
	RoleAdmin  = "admin"  // Administrador: gestiona restaurantes, menús y usuarios.
)

// User representa una cuenta registrada.
// El campo Password nunca se serializa a JSON (`json:"-"`); se almacena como hash bcrypt.
// Los tags `bson:` permiten que la misma estructura se persista tanto en Postgres como en Mongo
// sin necesidad de DTOs por adaptador.
type User struct {
	ID        string    `json:"id"         db:"id"         bson:"_id,omitempty"`
	Name      string    `json:"name"       db:"name"       bson:"name"`
	Email     string    `json:"email"      db:"email"      bson:"email"`
	Password  string    `json:"-"          db:"password"   bson:"password"`
	Role      string    `json:"role"       db:"role"       bson:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
}

// IsAdmin evita que los handlers comparen strings sueltos con "admin" por todos lados.
func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }
