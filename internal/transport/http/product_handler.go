package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// ProductHandler maneja las operaciones individuales sobre productos.
// La creación de productos se hace vía MenuHandler (POST /menus) porque los
// productos siempre pertenecen a un menú.
type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// GetByID godoc
// GET /products/:id
// Requiere: JWT
// Response 200: domain.Product
func (h *ProductHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

// ListByCategory godoc
// GET /products?category=:categoria
// Si no se pasa ?category devuelve todos los productos.
// Requiere: JWT
// Response 200: []domain.Product
func (h *ProductHandler) ListByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": "query param 'category' es requerido"})
		return
	}

	list, err := h.svc.ListByCategory(c.Request.Context(), category)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

// Update godoc
// PATCH /products/:id
// Requiere: JWT + role=admin
// Body: campos opcionales a modificar (name, description, category, price, available)
// Response 200: domain.Product actualizado
func (h *ProductHandler) Update(c *gin.Context) {
	role := c.GetString("role")
	id := c.Param("id")

	// Primero cargamos el producto existente para hacer patch parcial
	existing, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		renderError(c, err)
		return
	}

	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Category    *string  `json:"category"`
		Price       *float64 `json:"price"`
		Available   *bool    `json:"available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	// Aplicar solo los campos presentes en el body (patch semántico)
	updated := *existing
	if req.Name != nil {
		updated.Name = *req.Name
	}
	if req.Description != nil {
		updated.Description = *req.Description
	}
	if req.Category != nil {
		updated.Category = *req.Category
	}
	if req.Price != nil {
		updated.Price = *req.Price
	}
	if req.Available != nil {
		updated.Available = *req.Available
	}

	result, err := h.svc.Update(c.Request.Context(), role, &updated)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete godoc
// DELETE /products/:id
// Requiere: JWT + role=admin
// Response 204: No Content
func (h *ProductHandler) Delete(c *gin.Context) {
	role := c.GetString("role")
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), role, id); err != nil {
		renderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Ensure ProductHandler satisfies the domain contract at compile time.
var _ = domain.Product{}
