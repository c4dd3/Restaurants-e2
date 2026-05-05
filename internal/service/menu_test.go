package service

import (
	"context"
	"errors"
	"testing"

	"restaurants-e2/internal/domain"
)

func newMenuSvc(menus *mockMenuRepo, rests *mockRestaurantRepo, prods *mockProductRepo, cache *mockCache) *MenuService {
	return NewMenuService(menus, rests, prods, cache)
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestMenuServiceCreateAdmin(t *testing.T) {
	rests := newMockRestaurantRepo()
	rests.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1", Name: "La Tica"}
	cache := newMockCache()
	svc := newMenuSvc(newMockMenuRepo(), rests, newMockProductRepo(), cache)

	m, err := svc.Create(context.Background(), domain.RoleAdmin, domain.CreateMenuRequest{
		RestaurantID: "rest-1",
		Name:         "Menú Principal",
		Description:  "Comida típica",
		Products: []domain.ProductRequest{
			{Name: "Gallo Pinto", Category: "plato fuerte", Price: 3500, Available: true},
			{Name: "Café", Category: "bebida", Price: 1200, Available: true},
		},
	})

	if err != nil {
		t.Fatalf("Create inesperado: %v", err)
	}
	if m.ID == "" {
		t.Fatal("menú sin ID")
	}
	if len(m.Products) != 2 {
		t.Errorf("esperados 2 productos, obtenidos %d", len(m.Products))
	}
	if len(cache.deletedKeys) == 0 {
		t.Error("Create no invalidó la cache")
	}
}

func TestMenuServiceCreateClientForbidden(t *testing.T) {
	svc := newMenuSvc(newMockMenuRepo(), newMockRestaurantRepo(), newMockProductRepo(), newMockCache())

	_, err := svc.Create(context.Background(), domain.RoleClient, domain.CreateMenuRequest{
		RestaurantID: "rest-1", Name: "Menú", Description: "",
	})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

func TestMenuServiceCreateRestaurantNotFound(t *testing.T) {
	svc := newMenuSvc(newMockMenuRepo(), newMockRestaurantRepo(), newMockProductRepo(), newMockCache())

	_, err := svc.Create(context.Background(), domain.RoleAdmin, domain.CreateMenuRequest{
		RestaurantID: "no-existe", Name: "Menú",
	})

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("esperado ErrValidation por restaurante inexistente, obtenido %v", err)
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestMenuServiceGetByID(t *testing.T) {
	menus := newMockMenuRepo()
	menus.menus["menu-1"] = &domain.Menu{ID: "menu-1", Name: "Carta Especial"}
	svc := newMenuSvc(menus, newMockRestaurantRepo(), newMockProductRepo(), newMockCache())

	m, err := svc.GetByID(context.Background(), "menu-1")
	if err != nil {
		t.Fatalf("GetByID inesperado: %v", err)
	}
	if m.Name != "Carta Especial" {
		t.Errorf("nombre esperado 'Carta Especial', obtenido %q", m.Name)
	}
}

func TestMenuServiceGetByIDNotFound(t *testing.T) {
	svc := newMenuSvc(newMockMenuRepo(), newMockRestaurantRepo(), newMockProductRepo(), newMockCache())

	_, err := svc.GetByID(context.Background(), "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestMenuServiceUpdateAdmin(t *testing.T) {
	menus := newMockMenuRepo()
	menus.menus["menu-1"] = &domain.Menu{ID: "menu-1", Name: "Viejo Nombre"}
	cache := newMockCache()
	svc := newMenuSvc(menus, newMockRestaurantRepo(), newMockProductRepo(), cache)

	m, err := svc.Update(context.Background(), domain.RoleAdmin, "menu-1",
		domain.UpdateMenuRequest{Name: "Nuevo Nombre"})

	if err != nil {
		t.Fatalf("Update inesperado: %v", err)
	}
	if m.Name != "Nuevo Nombre" {
		t.Errorf("nombre no actualizado: %q", m.Name)
	}
	if len(cache.deletedKeys) == 0 {
		t.Error("Update no invalidó la cache")
	}
}

func TestMenuServiceUpdateClientForbidden(t *testing.T) {
	svc := newMenuSvc(newMockMenuRepo(), newMockRestaurantRepo(), newMockProductRepo(), newMockCache())

	_, err := svc.Update(context.Background(), domain.RoleClient, "menu-1",
		domain.UpdateMenuRequest{Name: "X"})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestMenuServiceDeleteAdmin(t *testing.T) {
	menus := newMockMenuRepo()
	menus.menus["menu-1"] = &domain.Menu{ID: "menu-1", Name: "A Eliminar"}
	cache := newMockCache()
	svc := newMenuSvc(menus, newMockRestaurantRepo(), newMockProductRepo(), cache)

	if err := svc.Delete(context.Background(), domain.RoleAdmin, "menu-1"); err != nil {
		t.Fatalf("Delete inesperado: %v", err)
	}
	if _, exists := menus.menus["menu-1"]; exists {
		t.Fatal("menú sigue en el repo después de Delete")
	}
}

func TestMenuServiceDeleteClientForbidden(t *testing.T) {
	svc := newMenuSvc(newMockMenuRepo(), newMockRestaurantRepo(), newMockProductRepo(), newMockCache())

	err := svc.Delete(context.Background(), domain.RoleClient, "menu-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

// ── Cache HIT ─────────────────────────────────────────────────────────────────

// ── Rutas de error ────────────────────────────────────────────────────────────

func TestMenuServiceCreateRestaurantError(t *testing.T) {
	// FindByID del restaurante devuelve error inesperado (no ErrNotFound).
	rests := newMockRestaurantRepo()
	rests.findByIDErr = errors.New("bd caída")
	svc := newMenuSvc(newMockMenuRepo(), rests, newMockProductRepo(), newMockCache())

	_, err := svc.Create(context.Background(), domain.RoleAdmin, domain.CreateMenuRequest{
		RestaurantID: "rest-1", Name: "Menú",
	})
	if err == nil {
		t.Fatal("esperaba error de BD en FindByID restaurante, obtuvo nil")
	}
}

func TestMenuServiceCreateMenuRepoError(t *testing.T) {
	// menus.Create falla con error inesperado.
	rests := newMockRestaurantRepo()
	rests.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1"}
	menus := newMockMenuRepo()
	menus.createErr = errors.New("error al insertar menú")
	svc := newMenuSvc(menus, rests, newMockProductRepo(), newMockCache())

	_, err := svc.Create(context.Background(), domain.RoleAdmin, domain.CreateMenuRequest{
		RestaurantID: "rest-1", Name: "Menú",
	})
	if err == nil {
		t.Fatal("esperaba error de BD en menus.Create, obtuvo nil")
	}
}

func TestMenuServiceUpdateMenuNotFound(t *testing.T) {
	// menus.Update devuelve ErrNotFound (propagado desde el repo).
	svc := newMenuSvc(newMockMenuRepo(), newMockRestaurantRepo(), newMockProductRepo(), newMockCache())

	_, err := svc.Update(context.Background(), domain.RoleAdmin, "no-existe",
		domain.UpdateMenuRequest{Name: "X"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestMenuServiceDeleteMenuNotFound(t *testing.T) {
	// menus.Delete devuelve ErrNotFound (propagado desde el repo).
	svc := newMenuSvc(newMockMenuRepo(), newMockRestaurantRepo(), newMockProductRepo(), newMockCache())

	err := svc.Delete(context.Background(), domain.RoleAdmin, "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestMenuServiceGetByIDCacheHit(t *testing.T) {
	cache := newMockCache()
	cached := &domain.Menu{ID: "menu-1", Name: "Menú Cacheado"}
	cache.Set(context.Background(), "menus:id:menu-1", cached, 0) //nolint:errcheck

	svc := newMenuSvc(newMockMenuRepo(), newMockRestaurantRepo(), newMockProductRepo(), cache)

	m, err := svc.GetByID(context.Background(), "menu-1")
	if err != nil {
		t.Fatalf("GetByID cache hit inesperado: %v", err)
	}
	if m.Name != "Menú Cacheado" {
		t.Errorf("esperado 'Menú Cacheado', obtenido %q", m.Name)
	}
}
