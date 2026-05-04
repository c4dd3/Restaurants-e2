package repomongo

import (
	"context"
	"testing"

	"restaurants-e2/internal/domain"
)

// Orden mínima para confirmar que guarda items y total correctamente.
func TestOrderRepoMongoCreateAndFindByID(t *testing.T) {
	repos, cleanup := testMongoRepositories(t)
	defer cleanup()
	ctx := context.Background()

	o := &domain.Order{ID: "order-1", UserID: "user-1", RestaurantID: "rest-1", Total: 9000, Items: []domain.OrderItem{{ProductID: "prod-1", Quantity: 2, Price: 4500}}}
	if err := repos.Orders.Create(ctx, o); err != nil {
		t.Fatal(err)
	}

	found, err := repos.Orders.FindByID(ctx, "order-1")
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || found.Total != 9000 || len(found.Items) != 1 {
		t.Fatalf("orden incorrecta: %#v", found)
	}
}
