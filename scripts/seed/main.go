//go:build ignore

package main

// main.go — entrada del seed generator.
//
// Flags:
//
//   --restaurants int        (default 10)   cantidad de restaurantes a crear.
//   --menus-per int          (default 2)    menús por restaurante.
//   --products-per int       (default 10)   productos por menú.
//   --users int              (default 30)   usuarios "normales" de prueba.
//   --admin-email string     (default "admin@example.com") email del admin seed.
//   --admin-password string  (default "admin12345") password del admin.
//   --reset bool             (default false) TRUNCATE/drop antes de insertar.
//   --dry-run bool           (default false) genera pero NO inserta; imprime a stdout.
//   --llm-batch int          (default 5)    cuántos restaurantes pedir por prompt.
//
// Flujo principal:
//
//   func main() {
//     1. Parsear flags.
//     2. Cargar config (reutilizar config.Load del api-service).
//     3. ctx, cancel := signal.NotifyContext(...)
//     4. Conectar repos:
//          repos, closeRepos, err := wiring.NewRepositories(ctx, cfg)
//          defer closeRepos(ctx)
//     5. Si --reset → pedir confirmación → ejecutar clean().
//     6. Crear admin: hashear password con auth.HashPassword, Insert directo.
//     7. Generar usuarios:
//          users := llmGenerateUsers(ctx, 30)
//          for u := range users { repos.Users.Create(ctx, u) }
//     8. Generar restaurantes en batches:
//          for i := 0; i < numRestaurants; i += llmBatch {
//              rs := llmGenerateRestaurants(ctx, llmBatch)
//              for r := range rs { repos.Restaurants.Create(ctx, r) }
//          }
//     9. Para cada restaurante, generar menús y productos (un prompt por
//        menú pidiendo productos coherentes con el tipo del restaurante).
//    10. Generar reservaciones: sample aleatorio de (user, restaurant, date
//        dentro de los próximos 14 días).
//    11. Opcional: generar órdenes históricas.
//    12. Si cfg.SearchEnabled → trigger reindex llamando al HTTP endpoint
//        /search/reindex (o invocar el indexer directo si es más simple).
//    13. Imprimir resumen: N restaurantes, N menús, N productos, N users.
//
// Manejo de errores:
//   - Errores del LLM (rate limit, timeout) → retry exponencial hasta 3 veces.
//   - Errores de BD en un item → loguear, continuar con siguientes (no abortar
//     todo el seed por un duplicate email).
//   - Al final, imprimir tabla de "errores-no-fatales" para que el usuario
//     decida si re-ejecutar.
//
// Observabilidad:
//   - Progress bar simple (github.com/schollz/progressbar o just log periódico).
//   - Loguear cada batch: "restaurantes 5/20 insertados (ok=5, err=0)".
//
// Exit codes:
//   0 → éxito total.
//   1 → error fatal (config, conexión a BD, LLM totalmente caído).
//   2 → completado con errores parciales (>N fallos individuales).
