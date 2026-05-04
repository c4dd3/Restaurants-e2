package http

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// Orden básica: crear con un producto existente y luego consultar el detalle.
func TestOrderHandlerCreateAndGet(t *testing.T) {
	setupGin()
	rests := newMockRestaurantRepo()
	products := newMockProductRepo()
	orders := newMockOrderRepo()
	_ = rests.Create(nil, &domain.Restaurant{ID: "rest-1", Name: "Soda TEC", Capacity: 20})
	_ = products.Create(nil, &domain.Product{ID: "prod-1", RestaurantID: "rest-1", Name: "Pizza", Category: "pizzas", Price: 4500, Available: true})
	h := NewOrderHandler(service.NewOrderService(orders, products, rests))
	r := gin.New()
	r.POST("/orders", func(c *gin.Context) { c.Set("user_id", "user-1"); h.Create(c) })
	r.GET("/orders/:id", func(c *gin.Context) { c.Set("user_id", "user-1"); c.Set("role", domain.RoleClient); h.GetByID(c) })

	body := domain.CreateOrderRequest{RestaurantID: "rest-1", Items: []domain.OrderItemRequest{{ProductID: "prod-1", Quantity: 2}}}
	w := performJSON(r, http.MethodPost, "/orders", body)
	requireStatus(t, w, http.StatusCreated)
	if !strings.Contains(w.Body.String(), "9000") {
		t.Fatalf("total esperado no aparece: %s", w.Body.String())
	}

	var created domain.Order
	if err := ginBindingJSON(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	w = performJSON(r, http.MethodGet, "/orders/"+created.ID, nil)
	requireStatus(t, w, http.StatusOK)
}

func TestOrderHandlerErrors(t *testing.T) {
	setupGin()
	rests := newMockRestaurantRepo()
	products := newMockProductRepo()
	orders := newMockOrderRepo()
	h := NewOrderHandler(service.NewOrderService(orders, products, rests))

	r := gin.New()
	r.POST("/orders", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		h.Create(c)
	})
	r.GET("/orders/:id", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("role", domain.RoleClient)
		h.GetByID(c)
	})

	// Orden inválida.
	w := performJSON(r, http.MethodPost, "/orders", map[string]any{
		"restaurant_id": "",
	})
	requireStatus(t, w, http.StatusBadRequest)

	// Orden inexistente.
	w = performJSON(r, http.MethodGet, "/orders/no-existe", nil)
	requireStatus(t, w, http.StatusNotFound)
}
