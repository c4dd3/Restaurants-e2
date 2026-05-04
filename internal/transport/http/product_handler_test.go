package http

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// Validación sencilla del handler de productos: obtener, filtrar, actualizar y borrar.
func TestProductHandlerGetListUpdateDelete(t *testing.T) {
	setupGin()
	products := newMockProductRepo()
	_ = products.Create(nil, &domain.Product{ID: "prod-1", RestaurantID: "rest-1", MenuID: "menu-1", Name: "Pizza", Category: "pizzas", Price: 4500, Available: true})
	h := NewProductHandler(service.NewProductService(products, mockCache{}))
	r := gin.New()
	r.GET("/products", h.ListByCategory)
	r.GET("/products/:id", h.GetByID)
	r.PATCH("/products/:id", func(c *gin.Context) { c.Set("role", domain.RoleAdmin); h.Update(c) })
	r.DELETE("/products/:id", func(c *gin.Context) { c.Set("role", domain.RoleAdmin); h.Delete(c) })

	w := performJSON(r, http.MethodGet, "/products/prod-1", nil)
	requireStatus(t, w, http.StatusOK)

	w = performJSON(r, http.MethodGet, "/products?category=pizzas", nil)
	requireStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "Pizza") {
		t.Fatalf("no encontró producto por categoría: %s", w.Body.String())
	}

	w = performJSON(r, http.MethodPatch, "/products/prod-1", map[string]any{"price": 5000})
	requireStatus(t, w, http.StatusOK)

	w = performJSON(r, http.MethodDelete, "/products/prod-1", nil)
	requireStatus(t, w, http.StatusNoContent)
}
