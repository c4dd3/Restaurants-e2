package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// MenuHandler maneja operaciones sobre menús y sus productos.
type MenuHandler struct {
	svc *service.MenuService
}

func NewMenuHandler(svc *service.MenuService) *MenuHandler {
	return &MenuHandler{svc: svc}
}

// Create godoc
// POST /menus
// Requiere: JWT + role=admin
// Body: { restaurant_id, name, description?, products[] }
// Response 201: domain.Menu con productos creados
func (h *MenuHandler) Create(c *gin.Context) {
	role := c.GetString("role")

	var req domain.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	menu, err := h.svc.Create(c.Request.Context(), role, req)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, menu)
}

// GetByID godoc
// GET /menus/:id
// Requiere: JWT
// Response 200: domain.Menu con sus productos
func (h *MenuHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	menu, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, menu)
}

// Update godoc
// PUT /menus/:id
// Requiere: JWT + role=admin
// Body: { name?, description?, products[]? }
// Si products viene en el body, reemplaza TODOS los productos del menú (TX atómica).
// Response 200: domain.Menu actualizado
func (h *MenuHandler) Update(c *gin.Context) {
	role := c.GetString("role")
	id := c.Param("id")

	var req domain.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	menu, err := h.svc.Update(c.Request.Context(), role, id, req)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, menu)
}

// Delete godoc
// DELETE /menus/:id
// Requiere: JWT + role=admin
// Los productos del menú se eliminan en cascada (ON DELETE CASCADE en Postgres).
// Response 204: No Content
func (h *MenuHandler) Delete(c *gin.Context) {
	role := c.GetString("role")
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), role, id); err != nil {
		renderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
