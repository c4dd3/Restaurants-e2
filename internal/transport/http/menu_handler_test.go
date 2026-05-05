package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// Este test recorre el CRUD principal del menú con un producto incluido.
func TestMenuHandlerCreateGetUpdateDelete(t *testing.T) {
	setupGin()
	rests := newMockRestaurantRepo()
	menus := newMockMenuRepo()
	products := newMockProductRepo()
	_ = rests.Create(nil, &domain.Restaurant{ID: "rest-1", Name: "Soda TEC", Capacity: 20})
	h := NewMenuHandler(service.NewMenuService(menus, rests, products, mockCache{}))
	r := gin.New()
	r.POST("/menus", func(c *gin.Context) { c.Set("role", domain.RoleAdmin); h.Create(c) })
	r.GET("/menus/:id", h.GetByID)
	r.PUT("/menus/:id", func(c *gin.Context) { c.Set("role", domain.RoleAdmin); h.Update(c) })
	r.DELETE("/menus/:id", func(c *gin.Context) { c.Set("role", domain.RoleAdmin); h.Delete(c) })

	createBody := domain.CreateMenuRequest{RestaurantID: "rest-1", Name: "Menú almuerzo", Products: []domain.ProductRequest{{Name: "Casado", Category: "almuerzos", Price: 3500, Available: true}}}
	w := performJSON(r, http.MethodPost, "/menus", createBody)
	requireStatus(t, w, http.StatusCreated)
	if !strings.Contains(w.Body.String(), "Casado") {
		t.Fatalf("menú no contiene producto: %s", w.Body.String())
	}

	var created domain.Menu
	if err := ginBindingJSON(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	w = performJSON(r, http.MethodGet, "/menus/"+created.ID, nil)
	requireStatus(t, w, http.StatusOK)

	w = performJSON(r, http.MethodPut, "/menus/"+created.ID, domain.UpdateMenuRequest{Name: "Menú actualizado"})
	requireStatus(t, w, http.StatusOK)

	w = performJSON(r, http.MethodDelete, "/menus/"+created.ID, nil)
	requireStatus(t, w, http.StatusNoContent)
}

func TestMenuHandlerErrors(t *testing.T) {
	setupGin()
	rests := newMockRestaurantRepo()
	menus := newMockMenuRepo()
	products := newMockProductRepo()

	h := NewMenuHandler(service.NewMenuService(menus, rests, products, mockCache{}))

	r := gin.New()
	r.POST("/menus", func(c *gin.Context) {
		c.Set("role", domain.RoleAdmin)
		h.Create(c)
	})
	r.GET("/menus/:id", h.GetByID)
	r.PUT("/menus/:id", func(c *gin.Context) {
		c.Set("role", domain.RoleAdmin)
		h.Update(c)
	})
	r.DELETE("/menus/:id", func(c *gin.Context) {
		c.Set("role", domain.RoleAdmin)
		h.Delete(c)
	})

	// Menú inexistente.
	w := performJSON(r, http.MethodGet, "/menus/no-existe", nil)
	requireStatus(t, w, http.StatusNotFound)

	// Create inválido.
	w = performJSON(r, http.MethodPost, "/menus", map[string]any{
		"name": "",
	})
	requireStatus(t, w, http.StatusBadRequest)

	// Update con JSON malo.
	req := httptest.NewRequest(http.MethodPut, "/menus/no-existe", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	requireStatus(t, w, http.StatusBadRequest)
}
