package main

import "fmt"

const systemPrompt = `Eres un generador de datos sintéticos para una app de restaurantes en Costa Rica.
Devuelves SIEMPRE JSON válido, sin markdown ni bloques de código.
Los nombres y direcciones deben ser realistas pero ficticios.`

func promptRestaurants(n, batchID int) (system, user string) {
	return systemPrompt, fmt.Sprintf(`Genera %d restaurantes costarricenses. batch-id: %d
Devuelve SOLO este JSON (sin texto extra):
{
  "restaurants": [
    {
      "name": "string (2-4 palabras, creativo)",
      "description": "string (1-2 oraciones sobre el restaurante)",
      "address": "string (dirección realista con provincia de Costa Rica)",
      "phone": "string (formato +506 XXXX-XXXX)",
      "category": "string (una de: tica|italiana|mexicana|japonesa|mariscos|vegetariana|parrilla|postres|cafetería)",
      "capacity": "number (entre 20 y 120)"
    }
  ]
}
IMPORTANTE: nombres únicos, variedad de provincias (San José, Heredia, Alajuela, Cartago, Guanacaste, Limón, Puntarenas).`, n, batchID)
}

func promptMenuWithProducts(restaurantName, category string, numProducts int) (system, user string) {
	return systemPrompt, fmt.Sprintf(`El restaurante "%s" es de categoría "%s".
Genera UN menú con exactamente %d productos coherentes con esa categoría.
Devuelve SOLO este JSON (sin texto extra):
{
  "menu": {
    "name": "string (ej: Menú Principal, Especialidades del Chef, Carta de Temporada)",
    "description": "string (1 oración describiendo el menú)",
    "products": [
      {
        "name": "string (nombre del plato)",
        "description": "string (descripción apetitosa de 1 oración)",
        "category": "string (una de: entrada|plato fuerte|bebida|postre|snack)",
        "price": "number (en colones CRC, entre 2500 y 18000)",
        "available": true
      }
    ]
  }
}
IMPORTANTE: precios realistas para Costa Rica, nombres creativos y variados.`, restaurantName, category, numProducts)
}

func promptUsers(n, batchID int) (system, user string) {
	return systemPrompt, fmt.Sprintf(`Genera %d usuarios costarricenses. batch-id: %d
Devuelve SOLO este JSON (sin texto extra):
{
  "users": [
    {
      "name": "string (nombre y apellido latino, realista)",
      "email": "string (email realista y único, lowercase)"
    }
  ]
}
IMPORTANTE: emails únicos, nombres variados con apellidos costarricenses.`, n, batchID)
}
