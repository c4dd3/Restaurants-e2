package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// OrderHandler maneja operaciones sobre pedidos.
type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Create godoc
// POST /orders
// Requiere: JWT
// Body: { restaurant_id, reservation_id?, items[{product_id, quantity}], pickup? }
// El total se calcula en el servidor — el cliente no lo envía.
// Response 201: domain.Order con total calculado e items con precio congelado
func (h *OrderHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	order, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetByID godoc
// GET /orders/:id
// Requiere: JWT
// Solo el dueño del pedido o un admin puede verlo (el service lo verifica).
// Response 200: domain.Order con items
func (h *OrderHandler) GetByID(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("role")
	id := c.Param("id")

	order, err := h.svc.GetByID(c.Request.Context(), userID, role, id)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, order)
}
