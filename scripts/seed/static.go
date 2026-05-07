package main

import (
	"context"
	"fmt"
	"math/rand"

	"restaurants-e2/internal/domain"
)

// StaticClient implementa LLMClient con datos hardcodeados.
// Se usa como fallback cuando ANTHROPIC_API_KEY no está disponible.
// El contrato es idéntico al AnthropicClient — el resto del seed no nota la diferencia.
type StaticClient struct{}

func (s *StaticClient) Complete(_ context.Context, _, user string) (string, error) {
	// Identificamos el tipo de prompt por palabras clave en el user message
	switch {
	case contains(user, "restaurantes costarricenses"):
		return staticRestaurantsJSON(), nil
	case contains(user, "usuarios costarricenses"):
		return staticUsersJSON(), nil
	default:
		// Prompt de menú con productos — extraemos nombre del restaurante del mensaje
		return staticMenuJSON(), nil
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ── Datos estáticos ──────────────────────────────────────────────────────────

var staticRestaurants = []domain.Restaurant{
	{Name: "La Soda de la Abuela", Address: "Barrio Escalante, San José", Phone: "+506 2222-1111", Description: "Comida casera costarricense con recetas tradicionales.", Capacity: 40},
	{Name: "El Rancho Tico", Address: "Alajuela Centro, Alajuela", Phone: "+506 2441-2222", Description: "Parrilladas y casados en ambiente familiar.", Capacity: 60},
	{Name: "Mariscos del Pacífico", Address: "Jacó, Puntarenas", Phone: "+506 2643-3333", Description: "Mariscos frescos traídos directamente del mar.", Capacity: 80},
	{Name: "Café Britt Terrace", Address: "Heredia Centro, Heredia", Phone: "+506 2260-4444", Description: "Café gourmet y repostería artesanal costarricense.", Capacity: 35},
	{Name: "Sushi Tico Fusion", Address: "Escazú, San José", Phone: "+506 2288-5555", Description: "Fusión japonesa con ingredientes locales costarricenses.", Capacity: 45},
	{Name: "La Cantina del Volcán", Address: "La Fortuna, Alajuela", Phone: "+506 2479-6666", Description: "Cocina tica con vista al Arenal.", Capacity: 70},
	{Name: "Sabor Caribeño", Address: "Puerto Limón, Limón", Phone: "+506 2758-7777", Description: "Cocina caribeña tradicional con arroz con pollo y patacones.", Capacity: 50},
	{Name: "Trattoria Bella Napoli", Address: "Curridabat, San José", Phone: "+506 2271-8888", Description: "Pasta y pizzas artesanales al horno de leña.", Capacity: 55},
	{Name: "Tacos y Más CR", Address: "Cartago Centro, Cartago", Phone: "+506 2551-9999", Description: "Cocina mexicana auténtica con ingredientes frescos.", Capacity: 40},
	{Name: "Verde Orgánico", Address: "Santa Ana, San José", Phone: "+506 2282-0000", Description: "Restaurante vegetariano y vegano con productos orgánicos locales.", Capacity: 30},
}

var staticUsers = []struct{ Name, Email string }{
	{"Carlos Rodríguez Mora", "carlos.rodriguez@gmail.com"},
	{"María Fernández Jiménez", "maria.fernandez@hotmail.com"},
	{"Andrés Quesada Vargas", "andres.quesada@yahoo.com"},
	{"Laura Solís Bermúdez", "laura.solis@gmail.com"},
	{"Diego Araya Brenes", "diego.araya@outlook.com"},
	{"Valentina Mora Castillo", "valentina.mora@gmail.com"},
	{"Sebastián Jiménez Rojas", "sebastian.jimenez@gmail.com"},
	{"Camila Herrera Ulate", "camila.herrera@hotmail.com"},
	{"José Ramírez Ugalde", "jose.ramirez@gmail.com"},
	{"Sofía Castro Méndez", "sofia.castro@yahoo.com"},
	{"Daniel Vega Porras", "daniel.vega@gmail.com"},
	{"Isabella Núñez Fonseca", "isabella.nunez@gmail.com"},
	{"Mateo Gutiérrez Alvarado", "mateo.gutierrez@hotmail.com"},
	{"Valeria Chaves Quirós", "valeria.chaves@gmail.com"},
	{"Alejandro Bolaños Arias", "alejandro.bolanos@outlook.com"},
	{"Natalia Vindas Salas", "natalia.vindas@gmail.com"},
	{"Francisco Umaña Picado", "francisco.umana@gmail.com"},
	{"Daniela Zamora Lobo", "daniela.zamora@hotmail.com"},
	{"Esteban Calvo Moreira", "esteban.calvo@gmail.com"},
	{"Paola Aguilar Cordero", "paola.aguilar@yahoo.com"},
}

var staticMenus = []struct {
	Name        string
	Description string
	Products    []domain.Product
}{
	{
		Name:        "Menú Principal",
		Description: "Nuestros platos más populares.",
		Products: []domain.Product{
			{Name: "Casado de pollo", Description: "Arroz, frijoles, ensalada y pollo asado.", Category: "plato fuerte", Price: 4500, Available: true},
			{Name: "Casado de carne", Description: "Arroz, frijoles, ensalada y bistec.", Category: "plato fuerte", Price: 5500, Available: true},
			{Name: "Gallo pinto", Description: "Arroz y frijoles salteados con natilla y huevo.", Category: "plato fuerte", Price: 2800, Available: true},
			{Name: "Refresco de cas", Description: "Refresco natural de cas con hielo.", Category: "bebida", Price: 1200, Available: true},
			{Name: "Fresco de mora", Description: "Refresco natural de mora con hielo.", Category: "bebida", Price: 1200, Available: true},
			{Name: "Flan de coco", Description: "Postre tradicional de coco con caramelo.", Category: "postre", Price: 1800, Available: true},
			{Name: "Tres leches", Description: "Pastel húmedo bañado en tres tipos de leche.", Category: "postre", Price: 2000, Available: true},
			{Name: "Chifrijo", Description: "Chicharrones, frijoles, pico de gallo y arroz.", Category: "entrada", Price: 3500, Available: true},
		},
	},
	{
		Name:        "Especialidades del Chef",
		Description: "Creaciones únicas de nuestra cocina.",
		Products: []domain.Product{
			{Name: "Ceviche de corvina", Description: "Corvina fresca marinada en limón con culantro.", Category: "entrada", Price: 4800, Available: true},
			{Name: "Filete de tilapia", Description: "Tilapia al ajillo con vegetales salteados.", Category: "plato fuerte", Price: 6500, Available: true},
			{Name: "Arroz con camarones", Description: "Arroz cremoso con camarones al ajillo.", Category: "plato fuerte", Price: 7500, Available: true},
			{Name: "Limonada natural", Description: "Limonada fresca con hielo y menta.", Category: "bebida", Price: 1500, Available: true},
			{Name: "Café chorreado", Description: "Café costarricense tradicional servido en chorreador.", Category: "bebida", Price: 1000, Available: true},
			{Name: "Churros con chocolate", Description: "Churros crujientes con salsa de chocolate caliente.", Category: "postre", Price: 2500, Available: true},
			{Name: "Empanadas de chiverre", Description: "Empanadas dulces con relleno de chiverre.", Category: "postre", Price: 1500, Available: true},
			{Name: "Patacones con guacamole", Description: "Plátano verde frito aplastado con guacamole fresco.", Category: "entrada", Price: 3000, Available: true},
		},
	},
}

// ── Generadores de JSON ──────────────────────────────────────────────────────

func staticRestaurantsJSON() string {
	items := ""
	for i, r := range staticRestaurants {
		comma := ","
		if i == len(staticRestaurants)-1 {
			comma = ""
		}
		items += fmt.Sprintf(`{"name":%q,"description":%q,"address":%q,"phone":%q,"category":"tica","capacity":%d}%s`,
			r.Name, r.Description, r.Address, r.Phone, r.Capacity, comma)
	}
	return `{"restaurants":[` + items + `]}`
}

func staticUsersJSON() string {
	items := ""
	for i, u := range staticUsers {
		comma := ","
		if i == len(staticUsers)-1 {
			comma = ""
		}
		items += fmt.Sprintf(`{"name":%q,"email":%q}%s`, u.Name, u.Email, comma)
	}
	return `{"users":[` + items + `]}`
}

func staticMenuJSON() string {
	m := staticMenus[rand.Intn(len(staticMenus))]
	products := ""
	for i, p := range m.Products {
		comma := ","
		if i == len(m.Products)-1 {
			comma = ""
		}
		products += fmt.Sprintf(`{"name":%q,"description":%q,"category":%q,"price":%.0f,"available":true}%s`,
			p.Name, p.Description, p.Category, p.Price, comma)
	}
	return fmt.Sprintf(`{"menu":{"name":%q,"description":%q,"products":[%s]}}`,
		m.Name, m.Description, products)
}
