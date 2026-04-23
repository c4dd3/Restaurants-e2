// Package http contiene los handlers Gin y los middlewares HTTP.
//
// Responsabilidad:
//   - Traducir HTTP → dominio: parsear DTOs de entrada con c.ShouldBindJSON.
//   - Invocar al service correspondiente.
//   - Traducir dominio → HTTP: elegir status code según el error devuelto,
//     serializar la respuesta a JSON.
//
// NO responsabilidad:
//   - Lógica de negocio (eso vive en service/).
//   - Acceso a BD (eso vive en adapters/).
//
// Estructura:
//
//	transport/http/
//	├── doc.go                  ← esto
//	├── router.go               ← NewRouter(services) *gin.Engine
//	├── errors.go               ← domainToHTTP(err) (status, payload)
//	├── middleware.go           ← AuthMiddleware, CORS, Logger custom
//	├── health_handler.go       ← GET /health
//	├── auth_handler.go         ← /auth/register, /auth/login
//	├── user_handler.go         ← /users/me
//	├── restaurant_handler.go   ← /restaurants (CRUD)
//	├── menu_handler.go         ← /menus (CRUD)
//	├── product_handler.go      ← /products (CRUD)
//	├── reservation_handler.go  ← /reservations
//	├── order_handler.go        ← /orders
//	└── search_handler.go       ← /search/products, /search/reindex (usado por search-service)
//
// Convenciones de handlers:
//   - Firma:  func (h *FooHandler) Create(c *gin.Context)
//   - Bind:   var req domain.CreateFooRequest; c.ShouldBindJSON(&req)
//     → si falla → c.JSON(400, gin.H{"error": err.Error()}) y return.
//   - Auth:   leer user_id y role del contexto (los puso AuthMiddleware).
//   - Call:   h.svc.DoThing(c.Request.Context(), args...)
//   - Error:  status, payload := domainToHTTP(err); c.JSON(status, payload).
//   - OK:     c.JSON(200, result).
package http
