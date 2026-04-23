// Package service contiene la lógica de negocio (capa de aplicación / casos de uso).
//
// Rol dentro de la arquitectura hexagonal:
//
//	transport/http  →  service  →  ports  ←  adapters
//
// El service ORQUESTA casos de uso: valida reglas de negocio, decide qué
// persistir, cuándo invalidar caché, si hay que rechazar por permisos, etc.
//
// Reglas que el service DEBE respetar:
//
//  1. NUNCA importa paquetes de adapters/ — solo domain/ y ports/.
//  2. NUNCA toca HTTP: no conoce gin.Context ni http.ResponseWriter.
//  3. Recibe context.Context en cada método público (cancelación/timeouts).
//  4. Los errores que devuelve son de dominio (ErrNotFound, ErrForbidden, etc.).
//     La capa transport los traduce a HTTP (404, 403, 409, ...).
//  5. Testeable en aislamiento con mocks de ports — sin BD ni red.
//
// Convención de constructores:
//
//	func NewRestaurantService(repo ports.RestaurantRepository, cache ports.Cache) *RestaurantService
//
// Cada service recibe SOLO las interfaces que necesita — no un struct gordo
// con todo. Eso ayuda a Interface Segregation (la "I" de SOLID).
package service
