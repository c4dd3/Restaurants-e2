package repopg

// db_errors_test.go — cubre todos los fmt.Errorf / pgErr que se ejecutan cuando
// la BD no está disponible.  Estrategia: obtener un pool válido de testPool() y
// cerrarlo de inmediato; a partir de ese punto cada operación retorna un error de
// conexión, ejerciendo las ramas que los happy-path tests nunca tocan.
//
// También cubre la rama fmt.Errorf("parsear DSN") de pool.go con un DSN
// cuyo puerto es no-numérico, lo que hace fallar pgxpool.ParseConfig.

import (
	"context"
	"testing"
	"time"

	"restaurants-e2/internal/config"
	"restaurants-e2/internal/domain"
)

// TestRepoPgDBErrorBranches dispara los errores de DB en todas las funciones del
// paquete usando un pool ya cerrado.  Cada sub-bloque cubre la(s) rama(s) de
// fmt.Errorf / pgErr que normalmente quedan en 0% con tests de happy-path.
func TestRepoPgDBErrorBranches(t *testing.T) {
	pool := testPool(t) // salta el test completo si Postgres no está disponible
	pool.Close()        // cerrar inmediatamente → todas las ops retornan error
	ctx := context.Background()

	// ── user.go ──────────────────────────────────────────────────────────────
	userRepo := NewUserRepoPg(pool)

	// Delete: fmt.Errorf("delete user: %w", err)
	if err := userRepo.Delete(ctx, "x"); err == nil {
		t.Error("userRepo.Delete: esperaba error con pool cerrado")
	}
	// FindByID → queryOneUser: fmt.Errorf("query user: %w", err)
	if _, err := userRepo.FindByID(ctx, "x"); err == nil {
		t.Error("userRepo.FindByID: esperaba error con pool cerrado")
	}

	// ── restaurant.go ────────────────────────────────────────────────────────
	restRepo := NewRestaurantRepoPg(pool)

	// Create: pgErr(err) desde QueryRow().Scan()
	if err := restRepo.Create(ctx, &domain.Restaurant{
		ID: "x", Name: "x", AdminID: "x",
	}); err == nil {
		t.Error("restRepo.Create: esperaba error con pool cerrado")
	}
	// FindByID: fmt.Errorf("query restaurant: %w", err)
	if _, err := restRepo.FindByID(ctx, "x"); err == nil {
		t.Error("restRepo.FindByID: esperaba error con pool cerrado")
	}
	// FindAll: fmt.Errorf("query restaurants: %w", err)
	if _, err := restRepo.FindAll(ctx); err == nil {
		t.Error("restRepo.FindAll: esperaba error con pool cerrado")
	}

	// ── product.go ───────────────────────────────────────────────────────────
	productRepo := NewProductRepoPg(pool)

	// FindByID: fmt.Errorf("query product: %w", err)
	if _, err := productRepo.FindByID(ctx, "x"); err == nil {
		t.Error("productRepo.FindByID: esperaba error con pool cerrado")
	}
	// Create: pgErr(err) desde pool.Exec()
	if err := productRepo.Create(ctx, &domain.Product{
		ID: "x", Name: "x", Category: "x",
	}); err == nil {
		t.Error("productRepo.Create: esperaba error con pool cerrado")
	}
	// Update: fmt.Errorf("update product: %w", err)
	if err := productRepo.Update(ctx, &domain.Product{
		ID: "x", Name: "x", Category: "x",
	}); err == nil {
		t.Error("productRepo.Update: esperaba error con pool cerrado")
	}
	// Delete: fmt.Errorf("delete product: %w", err)
	if err := productRepo.Delete(ctx, "x"); err == nil {
		t.Error("productRepo.Delete: esperaba error con pool cerrado")
	}
	// collectProducts (helper): fmt.Errorf("query products: %w", err)
	const productQ = `SELECT id, menu_id, restaurant_id, name, description, category, price, available FROM products`
	if _, err := collectProducts(ctx, pool, productQ); err == nil {
		t.Error("collectProducts: esperaba error con pool cerrado")
	}

	// ── menu.go ──────────────────────────────────────────────────────────────
	menuRepo := NewMenuRepoPg(pool)

	// Create: pgErr(err) desde QueryRow().Scan()
	if err := menuRepo.Create(ctx, &domain.Menu{
		ID: "x", RestaurantID: "x", Name: "x",
	}); err == nil {
		t.Error("menuRepo.Create: esperaba error con pool cerrado")
	}
	// FindByID: fmt.Errorf("query menu: %w", err)
	if _, err := menuRepo.FindByID(ctx, "x"); err == nil {
		t.Error("menuRepo.FindByID: esperaba error con pool cerrado")
	}
	// Update: fmt.Errorf("begin tx: %w", err) — BeginTx falla con pool cerrado
	if _, err := menuRepo.Update(ctx, "x", &domain.UpdateMenuRequest{Name: "x"}); err == nil {
		t.Error("menuRepo.Update: esperaba error con pool cerrado")
	}
	// Delete: fmt.Errorf("delete menu: %w", err)
	if err := menuRepo.Delete(ctx, "x"); err == nil {
		t.Error("menuRepo.Delete: esperaba error con pool cerrado")
	}
	// fetchProductsByMenuID (helper): fmt.Errorf("query products: %w", err)
	// pool implementa la interfaz querier (tiene el método Query), así que
	// podemos pasarlo directamente al helper interno.
	if _, err := fetchProductsByMenuID(ctx, pool, "x"); err == nil {
		t.Error("fetchProductsByMenuID: esperaba error con pool cerrado")
	}

	// ── reservation.go ───────────────────────────────────────────────────────
	reservationRepo := NewReservationRepoPg(pool)

	// Create: pgErr(err) desde QueryRow().Scan()
	if err := reservationRepo.Create(ctx, &domain.Reservation{
		ID:        "x",
		RestaurantID: "x",
		UserID:    "x",
		Date:      time.Now().Add(24 * time.Hour),
		PartySize: 2,
		Status:    domain.StatusPending,
	}); err == nil {
		t.Error("reservationRepo.Create: esperaba error con pool cerrado")
	}
	// FindByID: fmt.Errorf("query reservation: %w", err)
	if _, err := reservationRepo.FindByID(ctx, "x"); err == nil {
		t.Error("reservationRepo.FindByID: esperaba error con pool cerrado")
	}
	// Cancel: fmt.Errorf("cancel reservation: %w", err)
	if err := reservationRepo.Cancel(ctx, "x"); err == nil {
		t.Error("reservationRepo.Cancel: esperaba error con pool cerrado")
	}
	// CheckAvailability: fmt.Errorf("query capacity: %w", err)
	if _, err := reservationRepo.CheckAvailability(ctx, "x", 1); err == nil {
		t.Error("reservationRepo.CheckAvailability: esperaba error con pool cerrado")
	}

	// ── order.go ─────────────────────────────────────────────────────────────
	orderRepo := NewOrderRepoPg(pool)

	// Create: fmt.Errorf("begin tx: %w", err) — BeginTx falla con pool cerrado
	if err := orderRepo.Create(ctx, &domain.Order{
		ID: "x", UserID: "x", RestaurantID: "x", Status: domain.StatusPending,
	}); err == nil {
		t.Error("orderRepo.Create: esperaba error con pool cerrado")
	}
	// FindByID: fmt.Errorf("query order: %w", err)
	if _, err := orderRepo.FindByID(ctx, "x"); err == nil {
		t.Error("orderRepo.FindByID: esperaba error con pool cerrado")
	}
	// fetchOrderItems (helper): fmt.Errorf("query order_items: %w", err)
	if _, err := fetchOrderItems(ctx, pool, "x"); err == nil {
		t.Error("fetchOrderItems: esperaba error con pool cerrado")
	}
}

// TestNewPoolInvalidDSN cubre la rama fmt.Errorf("parsear DSN") de pool.go.
// Un puerto no-numérico hace que pgxpool.ParseConfig falle antes de intentar
// conectarse, sin necesitar infraestructura externa.
func TestNewPoolInvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewPool(ctx, config.PostgresConfig{DSN: "host=localhost port=not_a_number"})
	if err == nil {
		t.Fatal("NewPool con puerto inválido debería retornar error de parseo")
	}
}
