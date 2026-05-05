package repopg

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
)

func TestOrderRepoPgCreateAndFindByID(t *testing.T) {
	pool := testPool(t)
	userRepo := NewUserRepoPg(pool)
	restRepo := NewRestaurantRepoPg(pool)
	productRepo := NewProductRepoPg(pool)
	repo := NewOrderRepoPg(pool)
	ctx := context.Background()

	// Fixtures: user → restaurant → menu → product
	admin := seedAdminUser(t, userRepo)
	rest := seedRestaurant(t, restRepo, admin.ID)

	menuID := uuid.NewString()
	if _, err := pool.Exec(ctx,
		"INSERT INTO menus (id, restaurant_id, name, description, created_at, updated_at) VALUES ($1,$2,$3,$4,NOW(),NOW())",
		menuID, rest.ID, "Menú Test", "",
	); err != nil {
		t.Fatalf("insertar menú: %v", err)
	}

	prod := &domain.Product{
		ID: uuid.NewString(), MenuID: menuID, RestaurantID: rest.ID,
		Name: "Pizza", Category: "plato fuerte", Price: 8000, Available: true,
	}
	if err := productRepo.Create(ctx, prod); err != nil {
		t.Fatalf("insertar producto: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM order_items WHERE product_id=$1", prod.ID)  //nolint:errcheck
		pool.Exec(ctx, "DELETE FROM orders WHERE restaurant_id=$1", rest.ID)    //nolint:errcheck
		pool.Exec(ctx, "DELETE FROM products WHERE id=$1", prod.ID)             //nolint:errcheck
		pool.Exec(ctx, "DELETE FROM menus WHERE id=$1", menuID)                 //nolint:errcheck
		pool.Exec(ctx, "DELETE FROM restaurants WHERE id=$1", rest.ID)          //nolint:errcheck
	})

	// Crear orden con un item
	o := &domain.Order{
		ID:           uuid.NewString(),
		UserID:       admin.ID,
		RestaurantID: rest.ID,
		Items: []domain.OrderItem{
			{ProductID: prod.ID, Quantity: 2, Price: prod.Price},
		},
		Total:  prod.Price * 2,
		Status: domain.StatusPending,
		Pickup: false,
	}

	if err := repo.Create(ctx, o); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if o.CreatedAt.IsZero() {
		t.Error("Create no llenó created_at")
	}

	// FindByID
	found, err := repo.FindByID(ctx, o.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Total != o.Total {
		t.Errorf("total esperado %.0f, obtenido %.0f", o.Total, found.Total)
	}
	if len(found.Items) != 1 {
		t.Errorf("esperado 1 item, obtenidos %d", len(found.Items))
	}
	if found.Items[0].Quantity != 2 {
		t.Errorf("cantidad esperada 2, obtenida %d", found.Items[0].Quantity)
	}
}

func TestOrderRepoPgFindByIDNotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewOrderRepoPg(pool)

	_, err := repo.FindByID(context.Background(), uuid.NewString())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}
