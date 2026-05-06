package repopg

// errorpaths_test.go — cubre las ramas fmt.Errorf que solo disparan cuando
// el pool/DB falla durante una petición. La técnica es simple: se usa un
// contexto pre-cancelado. Cuando ctx ya está cancelado, pgxpool falla al
// intentar adquirir una conexión antes de ejecutar cualquier query, por lo
// que pool.Query / pool.Exec / pool.BeginTx / pool.QueryRow(...).Scan(...)
// retornan error inmediatamente, ejecutando la rama de error correspondiente.
//
// Esto cubre todos los `return fmt.Errorf("query X: %w", err)` y
// `return fmt.Errorf("begin tx: %w", err)` que son inalcanzables con una BD
// sana, pero son rutas de producción reales (e.g. request cancelado mid-flight).

import (
	"context"
	"testing"
	"time"

	"restaurants-e2/internal/config"
	"restaurants-e2/internal/domain"
)

// TestRepoPgCancelledContextErrors dispara las ramas de error de query/exec/tx
// en todos los repositorios usando un contexto ya cancelado.
func TestRepoPgCancelledContextErrors(t *testing.T) {
	pool := testPool(t) // se salta si Postgres no está disponible

	// Pre-cancelar antes de cualquier llamada al pool.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("UserFindByID", func(t *testing.T) {
		_, err := NewUserRepoPg(pool).FindByID(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("UserFindByEmail", func(t *testing.T) {
		_, err := NewUserRepoPg(pool).FindByEmail(ctx, "x@x.com")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("UserCreate", func(t *testing.T) {
		err := NewUserRepoPg(pool).Create(ctx, &domain.User{
			ID: "x", Name: "X", Email: "x@x.com", Password: "h", Role: domain.RoleClient,
		})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("UserUpdate", func(t *testing.T) {
		_, err := NewUserRepoPg(pool).Update(ctx, "x", &domain.UpdateUserRequest{Name: "Y"})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("UserDelete", func(t *testing.T) {
		err := NewUserRepoPg(pool).Delete(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("RestaurantCreate", func(t *testing.T) {
		err := NewRestaurantRepoPg(pool).Create(ctx, &domain.Restaurant{
			ID: "x", Name: "X", Address: "X", Phone: "X", AdminID: "x", Capacity: 10,
		})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("RestaurantFindByID", func(t *testing.T) {
		_, err := NewRestaurantRepoPg(pool).FindByID(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("RestaurantFindAll", func(t *testing.T) {
		_, err := NewRestaurantRepoPg(pool).FindAll(ctx)
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("MenuCreate", func(t *testing.T) {
		err := NewMenuRepoPg(pool).Create(ctx, &domain.Menu{
			ID: "x", RestaurantID: "r", Name: "X",
		})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("MenuFindByID", func(t *testing.T) {
		_, err := NewMenuRepoPg(pool).FindByID(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("MenuUpdate", func(t *testing.T) {
		_, err := NewMenuRepoPg(pool).Update(ctx, "x", &domain.UpdateMenuRequest{Name: "Y"})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("MenuDelete", func(t *testing.T) {
		err := NewMenuRepoPg(pool).Delete(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ProductFindByID", func(t *testing.T) {
		_, err := NewProductRepoPg(pool).FindByID(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ProductFindByIDs", func(t *testing.T) {
		_, err := NewProductRepoPg(pool).FindByIDs(ctx, []string{"x"})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ProductFindByCategory", func(t *testing.T) {
		_, err := NewProductRepoPg(pool).FindByCategory(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ProductFindAll", func(t *testing.T) {
		_, err := NewProductRepoPg(pool).FindAll(ctx)
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ProductCreate", func(t *testing.T) {
		err := NewProductRepoPg(pool).Create(ctx, &domain.Product{
			ID: "x", MenuID: "m", RestaurantID: "r",
			Name: "X", Category: "X", Price: 100,
		})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ProductUpdate", func(t *testing.T) {
		err := NewProductRepoPg(pool).Update(ctx, &domain.Product{
			ID: "x", Name: "X", Category: "X", Price: 100,
		})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ProductDelete", func(t *testing.T) {
		err := NewProductRepoPg(pool).Delete(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ReservationCreate", func(t *testing.T) {
		err := NewReservationRepoPg(pool).Create(ctx, &domain.Reservation{
			ID: "x", RestaurantID: "r", UserID: "u",
			Date:      time.Now().Add(24 * time.Hour),
			PartySize: 2, Status: domain.StatusPending,
		})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ReservationFindByID", func(t *testing.T) {
		_, err := NewReservationRepoPg(pool).FindByID(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ReservationCancel", func(t *testing.T) {
		err := NewReservationRepoPg(pool).Cancel(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("ReservationCheckAvailability", func(t *testing.T) {
		_, err := NewReservationRepoPg(pool).CheckAvailability(ctx, "r", 2)
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("OrderCreate", func(t *testing.T) {
		err := NewOrderRepoPg(pool).Create(ctx, &domain.Order{
			ID: "x", UserID: "u", RestaurantID: "r", Status: domain.StatusPending,
		})
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})

	t.Run("OrderFindByID", func(t *testing.T) {
		_, err := NewOrderRepoPg(pool).FindByID(ctx, "x")
		if err == nil {
			t.Error("esperaba error por contexto cancelado")
		}
	})
}

// TestNewPoolInvalidDSNFormat cubre la rama `fmt.Errorf("parsear DSN: %w", err)`
// de NewPool: pgxpool.ParseConfig falla cuando la URL tiene un formato inválido
// (bracket abierto en el host → url.Parse retorna error).
func TestNewPoolInvalidDSNFormat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewPool(ctx, config.PostgresConfig{DSN: "postgresql://[invalid_host"})
	if err == nil {
		t.Fatal("esperaba error por DSN con URL inválida")
	}
}
