package http

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// Acá pruebo el flujo normal de restaurantes: crear, listar y consultar por id.
func TestRestaurantHandlerCreateListAndGet(t *testing.T) {
	setupGin()
	rests := newMockRestaurantRepo()
	h := NewRestaurantHandler(service.NewRestaurantService(rests, mockCache{}))
	r := gin.New()
	r.POST("/restaurants", func(c *gin.Context) { c.Set("user_id", "admin-1"); c.Set("role", domain.RoleAdmin); h.Create(c) })
	r.GET("/restaurants", h.List)
	r.GET("/restaurants/:id", h.GetByID)

	body := domain.CreateRestaurantRequest{Name: "Soda TEC", Address: "Cartago", Phone: "8888-8888", Description: "Comida casera", Capacity: 25}
	w := performJSON(r, http.MethodPost, "/restaurants", body)
	requireStatus(t, w, http.StatusCreated)
	if !strings.Contains(w.Body.String(), "Soda TEC") {
		t.Fatalf("no contiene restaurante creado: %s", w.Body.String())
	}

	var created domain.Restaurant
	if err := ginBindingJSON(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	w = performJSON(r, http.MethodGet, "/restaurants", nil)
	requireStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "Soda TEC") {
		t.Fatalf("listado no contiene restaurante: %s", w.Body.String())
	}

	w = performJSON(r, http.MethodGet, "/restaurants/"+created.ID, nil)
	requireStatus(t, w, http.StatusOK)
}
