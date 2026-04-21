package domain

// DefaultProductDescription es el texto que usa el servicio de búsqueda
// cuando un producto carece de descripción. El enunciado de Etapa 2 lo exige explícitamente.
const DefaultProductDescription = "Producto sin descripción"

// Product reemplaza al MenuItem de Etapa 1.
// Se llama "producto" para alinear el vocabulario con el enunciado (sharding y ElasticSearch).
// Campo Category es nuevo y es la clave del filtrado/indexado en Etapa 2.
//
// Nota sobre sharding en MongoDB:
// La shard key recomendada para esta colección es `category` (hashed) — balancea bien
// la distribución de escrituras cuando hay variedad de categorías, y permite consultas
// dirigidas a un solo shard al filtrar por categoría.
type Product struct {
	ID           string  `json:"id"            db:"id"            bson:"_id,omitempty"`
	MenuID       string  `json:"menu_id"       db:"menu_id"       bson:"menu_id"`
	RestaurantID string  `json:"restaurant_id" db:"restaurant_id" bson:"restaurant_id"`
	Name         string  `json:"name"          db:"name"          bson:"name"`
	Description  string  `json:"description"   db:"description"   bson:"description"`
	Category     string  `json:"category"      db:"category"      bson:"category"`
	Price        float64 `json:"price"         db:"price"         bson:"price"`
	Available    bool    `json:"available"     db:"available"     bson:"available"`
}

// DescriptionOrDefault devuelve la descripción, o el texto por defecto si está vacía.
// Encapsular esta regla acá evita que cada adaptador/handler la reimplemente.
func (p *Product) DescriptionOrDefault() string {
	if p.Description == "" {
		return DefaultProductDescription
	}
	return p.Description
}
