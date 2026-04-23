package http

// router.go — ensambla el *gin.Engine con todas las rutas del api-service.
//
// Función pública:
//
//   NewRouter(deps Deps) *gin.Engine
//
// Donde Deps agrupa lo que el router necesita inyectar a los handlers:
//
//   type Deps struct {
//       AuthService        *service.AuthService
//       UserService        *service.UserService
//       RestaurantService  *service.RestaurantService
//       MenuService        *service.MenuService
//       ProductService     *service.ProductService
//       ReservationService *service.ReservationService
//       OrderService       *service.OrderService
//       SearchService      ports.SearchIndex    // solo para /search/products (lectura)
//       JWTSecret          string               // para AuthMiddleware
//   }
//
// Construcción:
//
//   1. r := gin.New()
//   2. r.Use(gin.Recovery(), middleware.RequestID(), middleware.Logger(), middleware.CORS())
//   3. Registrar health:
//        r.GET("/health", HealthHandler)
//   4. Grupo público /auth:
//        auth := r.Group("/auth")
//        auth.POST("/register", ah.Register)
//        auth.POST("/login", ah.Login)
//   5. Grupo protegido (usa AuthMiddleware):
//        api := r.Group("/")
//        api.Use(AuthMiddleware(deps.JWTSecret))
//        api.GET("/users/me", uh.Me)
//        api.PATCH("/users/me", uh.Update)
//        api.POST("/restaurants", rh.Create)       // role=admin (validado en service)
//        api.GET("/restaurants", rh.List)
//        api.GET("/restaurants/:id", rh.GetByID)
//        ... etc para menus, products, reservations, orders, search.
//   6. return r
//
// Por qué un solo NewRouter y no múltiples:
//   - Centraliza el mapa de rutas; fácil de leer en un vistazo.
//   - Simplifica los tests de integración (un router == un http.Handler).
//   - Evita que cada handler se registre solo (efecto "spooky action").
//
// Por qué NO usar grupos por "module" (ej. r.Group("/restaurants")):
//   - Lo aceptable pero complica el paso de dependencias. Cada handler es
//     pequeño, con 2-4 rutas; inline queda legible y plano.
