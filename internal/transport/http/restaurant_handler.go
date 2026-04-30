package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// RestaurantHandler maneja operaciones sobre restaurantes.
type RestaurantHandler struct {
	svc *service.RestaurantService
}

func NewRestaurantHandler(svc *service.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{svc: svc}
}

// Create godoc
// POST /restaurants
// Requiere: JWT + role=admin (validado en el service)
// Body: { name, address, phone, description?, capacity }
// Response 201: domain.Restaurant
func (h *RestaurantHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("role")

	var req domain.CreateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	restaurant, err := h.svc.Create(c.Request.Context(), userID, role, req)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, restaurant)
}

// List godoc
// GET /restaurants
// Público (no requiere JWT)
// Response 200: []domain.Restaurant
func (h *RestaurantHandler) List(c *gin.Context) {
	restaurants, err := h.svc.List(c.Request.Context())
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": restaurants})
}

// GetByID godoc
// GET /restaurants/:id
// Público (no requiere JWT)
// Response 200: domain.Restaurant
func (h *RestaurantHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	restaurant, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, restaurant)
}
