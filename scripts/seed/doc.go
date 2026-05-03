//go:build ignore

// Package main (scripts/seed) — generador de datos de prueba asistido por LLM.
//
// Propósito: cumplir el requisito del enunciado "uso de LLMs para generar
// datos realistas". Este binario se ejecuta una vez (o N veces) contra un
// entorno de desarrollo para poblarlo con restaurantes, menús, productos,
// usuarios, reservaciones y órdenes creíbles.
//
// NO es parte del runtime de producción. Se invoca manualmente:
//
//   go run ./scripts/seed --restaurants=20 --menus-per=3 --products-per=12
//
// Estructura:
//
//   scripts/seed/
//   ├── doc.go          ← esto
//   ├── main.go         ← entrada, parseo de flags, loop de generación
//   ├── llm.go          ← cliente HTTP contra OpenAI o Anthropic
//   └── prompts.go      ← templates de prompt con few-shot examples
//
// Flujo:
//   1. Conectarse directamente a la BD primaria (postgres o mongo) usando
//      wiring.NewRepositories(ctx, cfg). NO va por HTTP — es más rápido y
//      evita autenticación JWT.
//   2. Llamar al LLM con un prompt tipo "Genera 20 restaurantes de comida
//      tica con nombre, descripción, dirección realista en Costa Rica..."
//   3. Parsear la respuesta JSON (pedir al LLM que devuelva JSON estructurado
//      con response_format o tool use).
//   4. Insertar en BD vía los repos inyectados.
//   5. Repetir para menús, productos, usuarios, reservaciones.
//
// Por qué conectar al repo directo y no al API:
//   - Velocidad: 100 inserts por HTTP = 100 round-trips + JWT. Por BD = batch.
//   - Simplicidad: no necesita un admin token previo.
//   - No contamina los logs de requests reales.
//
// Idempotencia:
//   - El script NO es idempotente por defecto (genera nuevos datos cada
//     corrida). Flag --reset ejecuta TRUNCATE (pg) o dropDatabase (mongo)
//     antes. Usar con mucho cuidado — requiere confirmación interactiva.
//
// Variables de entorno necesarias:
//   - DB_ENGINE, DATABASE_URL (o MONGO_URI) — como el api-service.
//   - OPENAI_API_KEY o ANTHROPIC_API_KEY — para el LLM.
//   - SEED_LLM_MODEL — por defecto "claude-haiku" o "gpt-4o-mini" (balance
//     costo/calidad para generar datos).
package main
