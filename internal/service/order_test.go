package service

import (
	"context"
	"errors"
	"testing"

	"restaurants-e2/internal/domain"
)

func newOrderSvc(orders *mockOrderRepo, prods *mockProductRepo, rests *mockRestaurantRepo) *OrderService {
	return NewOrderService(orders, prods, rests)
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestOrderServiceCreateHappyPath(t *testing.T) {
	rests := newMockRestaurantRepo()
	rests.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1"}

	prods := newMockProductRepo()
	prods.products["prod-1"] = &domain.Product{ID: "prod-1", RestaurantID: "rest-1", Price: 5000, Category: "plato fuerte"}
	prods.products["prod-2"] = &domain.Product{ID: "prod-2", RestaurantID: "rest-1", Price: 1500, Category: "bebida"}

	orders := newMockOrderRepo()
	svc := newOrderSvc(orders, prods, rests)

	o, err := svc.Create(context.Background(), "user-1", domain.CreateOrderRequest{
		RestaurantID: "rest-1",
		Items: []domain.OrderItemRequest{
			{ProductID: "prod-1", Quantity: 2},
			{ProductID: "prod-2", Quantity: 1},
		},
	})

	if err != nil {
		t.Fatalf("Create inesperado: %v", err)
	}
	if o.ID == "" {
		t.Fatal("orden sin ID")
	}
	// Total esperado: 5000*2 + 1500*1 = 11500
	if o.Total != 11500 {
		t.Errorf("total esperado 11500, obtenido %.0f", o.Total)
	}
	if len(o.Items) != 2 {
		t.Errorf("esperados 2 items, obtenidos %d", len(o.Items))
	}
	if o.Status != domain.StatusPending {
		t.Errorf("estado esperado pending, obtenido %q", o.Status)
	}
}

func TestOrderServiceCreateRestaurantNotFound(t *testing.T) {
	svc := newOrderSvc(newMockOrderRepo(), newMockProductRepo(), newMockRestaurantRepo())

	_, err := svc.Create(context.Background(), "user-1", domain.CreateOrderRequest{
		RestaurantID: "no-existe",
		Items:        []domain.OrderItemRequest{{ProductID: "p-1", Quantity: 1}},
	})

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("esperado ErrValidation, obtenido %v", err)
	}
}

func TestOrderServiceCreateProductNotFound(t *testing.T) {
	rests := newMockRestaurantRepo()
	rests.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1"}
	svc := newOrderSvc(newMockOrderRepo(), newMockProductRepo(), rests)

	_, err := svc.Create(context.Background(), "user-1", domain.CreateOrderRequest{
		RestaurantID: "rest-1",
		Items:        []domain.OrderItemRequest{{ProductID: "prod-inexistente", Quantity: 1}},
	})

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("esperado ErrValidation por producto inexistente, obtenido %v", err)
	}
}

func TestOrderServiceCreateProductWrongRestaurant(t *testing.T) {
	rests := newMockRestaurantRepo()
	rests.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1"}
	rests.restaurants["rest-2"] = &domain.Restaurant{ID: "rest-2"}

	prods := newMockProductRepo()
	// Producto pertenece a rest-2, pero la orden es de rest-1
	prods.products["prod-1"] = &domain.Product{ID: "prod-1", RestaurantID: "rest-2", Price: 3000}

	svc := newOrderSvc(newMockOrderRepo(), prods, rests)

	_, err := svc.Create(context.Background(), "user-1", domain.CreateOrderRequest{
		RestaurantID: "rest-1",
		Items:        []domain.OrderItemRequest{{ProductID: "prod-1", Quantity: 1}},
	})

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("esperado ErrValidation por producto de otro restaurante, obtenido %v", err)
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestOrderServiceGetByIDOwner(t *testing.T) {
	orders := newMockOrderRepo()
	orders.orders["ord-1"] = &domain.Order{ID: "ord-1", UserID: "user-1", Total: 8000}
	svc := newOrderSvc(orders, newMockProductRepo(), newMockRestaurantRepo())

	o, err := svc.GetByID(context.Background(), "user-1", domain.RoleClient, "ord-1")
	if err != nil {
		t.Fatalf("GetByID inesperado: %v", err)
	}
	if o.Total != 8000 {
		t.Errorf("total esperado 8000, obtenido %.0f", o.Total)
	}
}

func TestOrderServiceGetByIDAdminSeesAll(t *testing.T) {
	orders := newMockOrderRepo()
	orders.orders["ord-1"] = &domain.Order{ID: "ord-1", UserID: "user-99", Total: 5000}
	svc := newOrderSvc(orders, newMockProductRepo(), newMockRestaurantRepo())

	// Admin puede ver la orden de cualquier usuario
	_, err := svc.GetByID(context.Background(), "admin-1", domain.RoleAdmin, "ord-1")
	if err != nil {
		t.Fatalf("admin GetByID inesperado: %v", err)
	}
}

func TestOrderServiceGetByIDForbiddenOtherUser(t *testing.T) {
	orders := newMockOrderRepo()
	orders.orders["ord-1"] = &domain.Order{ID: "ord-1", UserID: "user-1"}
	svc := newOrderSvc(orders, newMockProductRepo(), newMockRestaurantRepo())

	_, err := svc.GetByID(context.Background(), "user-2", domain.RoleClient, "ord-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

func TestOrderServiceGetByIDNotFound(t *testing.T) {
	svc := newOrderSvc(newMockOrderRepo(), newMockProductRepo(), newMockRestaurantRepo())

	_, err := svc.GetByID(context.Background(), "user-1", domain.RoleClient, "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}
