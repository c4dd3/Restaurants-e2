package http

import (
	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/service"
)

// Deps agrupa todas las dependencias que el router necesita para construir los handlers.
// Se inyectan desde cmd/api/main.go tras el wiring de repos y servicios.
type Deps struct {
	AuthService        *service.AuthService
	UserService        *service.UserService
	RestaurantService  *service.RestaurantService
	MenuService        *service.MenuService
	ReservationService *service.ReservationService
	OrderService       *service.OrderService
	JWTSecret          string
}

// NewRouter construye el *gin.Engine con todas las rutas del api-service.
// Orden de middlewares: Recovery → RequestID → Logger de Gin.
func NewRouter(deps Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), RequestID())

	// ── Health ──────────────────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "api", "status": "ok"})
	})

	// ── Handlers ─────────────────────────────────────────────────────────────
	ah := NewAuthHandler(deps.AuthService)
	uh := NewUserHandler(deps.UserService)
	rh := NewRestaurantHandler(deps.RestaurantService)
	mh := NewMenuHandler(deps.MenuService)
	resh := NewReservationHandler(deps.ReservationService)
	oh := NewOrderHandler(deps.OrderService)

	// ── Rutas públicas (sin JWT) ──────────────────────────────────────────────
	auth := r.Group("/auth")
	{
		auth.POST("/register", ah.Register)
		auth.POST("/login", ah.Login)
	}

	// GET /restaurants y GET /restaurants/:id son públicos
	r.GET("/restaurants", rh.List)
	r.GET("/restaurants/:id", rh.GetByID)

	// ── Rutas protegidas (requieren JWT válido) ───────────────────────────────
	api := r.Group("/")
	api.Use(AuthMiddleware(deps.JWTSecret))
	{
		// Usuarios
		api.GET("/users/me", uh.GetMe)
		api.PUT("/users/:id", uh.Update)
		api.DELETE("/users/:id", uh.Delete)

		// Restaurantes (solo admin — el service verifica el rol)
		api.POST("/restaurants", rh.Create)

		// Menús (solo admin — el service verifica el rol)
		api.POST("/menus", mh.Create)
		api.GET("/menus/:id", mh.GetByID)
		api.PUT("/menus/:id", mh.Update)
		api.DELETE("/menus/:id", mh.Delete)

		// Reservas
		api.POST("/reservations", resh.Create)
		api.DELETE("/reservations/:id", resh.Cancel)

		// Órdenes
		api.POST("/orders", oh.Create)
		api.GET("/orders/:id", oh.GetByID)
	}

	return r
}
