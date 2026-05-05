package service

import (
	"context"
	"errors"
	"testing"

	"restaurants-e2/internal/domain"
)

func newProductSvc(prods *mockProductRepo, cache *mockCache) *ProductService {
	return NewProductService(prods, cache)
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestProductServiceGetByID(t *testing.T) {
	prods := newMockProductRepo()
	prods.products["p-1"] = &domain.Product{ID: "p-1", Name: "Casado", Category: "plato fuerte", Price: 5000}
	svc := newProductSvc(prods, newMockCache())

	p, err := svc.GetByID(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("GetByID inesperado: %v", err)
	}
	if p.Name != "Casado" {
		t.Errorf("nombre esperado 'Casado', obtenido %q", p.Name)
	}
}

func TestProductServiceGetByIDNotFound(t *testing.T) {
	svc := newProductSvc(newMockProductRepo(), newMockCache())

	_, err := svc.GetByID(context.Background(), "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

// ── ListByCategory ────────────────────────────────────────────────────────────

func TestProductServiceListByCategory(t *testing.T) {
	prods := newMockProductRepo()
	prods.products["p-1"] = &domain.Product{ID: "p-1", Name: "Gallo Pinto", Category: "plato fuerte"}
	prods.products["p-2"] = &domain.Product{ID: "p-2", Name: "Café", Category: "bebida"}
	prods.products["p-3"] = &domain.Product{ID: "p-3", Name: "Casado", Category: "plato fuerte"}
	svc := newProductSvc(prods, newMockCache())

	list, err := svc.ListByCategory(context.Background(), "plato fuerte")
	if err != nil {
		t.Fatalf("ListByCategory inesperado: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("esperados 2 productos, obtenidos %d", len(list))
	}
}

func TestProductServiceListByCategoryEmpty(t *testing.T) {
	svc := newProductSvc(newMockProductRepo(), newMockCache())

	list, err := svc.ListByCategory(context.Background(), "categoria-inexistente")
	if err != nil {
		t.Fatalf("ListByCategory inesperado: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("esperada lista vacía, obtenidos %d", len(list))
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestProductServiceUpdateAdmin(t *testing.T) {
	prods := newMockProductRepo()
	prods.products["p-1"] = &domain.Product{ID: "p-1", Name: "Viejo", Category: "postre", Price: 2000}
	cache := newMockCache()
	svc := newProductSvc(prods, cache)

	updated, err := svc.Update(context.Background(), domain.RoleAdmin,
		&domain.Product{ID: "p-1", Name: "Nuevo", Category: "postre", Price: 2500})

	if err != nil {
		t.Fatalf("Update inesperado: %v", err)
	}
	if updated.Name != "Nuevo" || updated.Price != 2500 {
		t.Errorf("producto no actualizado correctamente: %+v", updated)
	}
	if len(cache.deletedKeys) == 0 {
		t.Error("Update no invalidó la cache")
	}
}

func TestProductServiceUpdateClientForbidden(t *testing.T) {
	svc := newProductSvc(newMockProductRepo(), newMockCache())

	_, err := svc.Update(context.Background(), domain.RoleClient,
		&domain.Product{ID: "p-1", Name: "X"})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestProductServiceDeleteAdmin(t *testing.T) {
	prods := newMockProductRepo()
	prods.products["p-1"] = &domain.Product{ID: "p-1", Name: "A eliminar"}
	cache := newMockCache()
	svc := newProductSvc(prods, cache)

	if err := svc.Delete(context.Background(), domain.RoleAdmin, "p-1"); err != nil {
		t.Fatalf("Delete inesperado: %v", err)
	}
	if _, exists := prods.products["p-1"]; exists {
		t.Fatal("producto sigue en el repo después de Delete")
	}
	if len(cache.deletedKeys) == 0 {
		t.Error("Delete no invalidó la cache")
	}
}

func TestProductServiceDeleteClientForbidden(t *testing.T) {
	svc := newProductSvc(newMockProductRepo(), newMockCache())

	err := svc.Delete(context.Background(), domain.RoleClient, "p-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

// ── Cache HIT paths ───────────────────────────────────────────────────────────

// ── Rutas de error ────────────────────────────────────────────────────────────

func TestProductServiceListByCategoryError(t *testing.T) {
	prods := newMockProductRepo()
	prods.findByCategoryErr = errors.New("bd caída")
	svc := newProductSvc(prods, newMockCache())

	_, err := svc.ListByCategory(context.Background(), "plato fuerte")
	if err == nil {
		t.Fatal("esperaba error de BD en FindByCategory, obtuvo nil")
	}
}

func TestProductServiceUpdateNotFound(t *testing.T) {
	// mockProductRepo.Update devuelve ErrNotFound para productos inexistentes.
	svc := newProductSvc(newMockProductRepo(), newMockCache())

	_, err := svc.Update(context.Background(), domain.RoleAdmin,
		&domain.Product{ID: "no-existe", Name: "X", Category: "postre", Price: 100})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestProductServiceDeleteNotFound(t *testing.T) {
	// mockProductRepo.Delete devuelve ErrNotFound para productos inexistentes.
	svc := newProductSvc(newMockProductRepo(), newMockCache())

	err := svc.Delete(context.Background(), domain.RoleAdmin, "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

func TestProductServiceGetByIDCacheHit(t *testing.T) {
	cache := newMockCache()
	cached := &domain.Product{ID: "p-1", Name: "Producto Cacheado", Category: "postre", Price: 1500}
	cache.Set(context.Background(), "products:id:p-1", cached, 0) //nolint:errcheck

	svc := newProductSvc(newMockProductRepo(), cache)

	p, err := svc.GetByID(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("GetByID cache hit inesperado: %v", err)
	}
	if p.Name != "Producto Cacheado" {
		t.Errorf("esperado 'Producto Cacheado', obtenido %q", p.Name)
	}
}

func TestProductServiceListByCategoryCacheHit(t *testing.T) {
	cache := newMockCache()
	cachedList := []domain.Product{
		{ID: "p-1", Name: "Gallo Pinto", Category: "plato fuerte"},
	}
	cache.Set(context.Background(), "products:cat:plato fuerte", cachedList, 0) //nolint:errcheck

	svc := newProductSvc(newMockProductRepo(), cache)

	list, err := svc.ListByCategory(context.Background(), "plato fuerte")
	if err != nil {
		t.Fatalf("ListByCategory cache hit inesperado: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("esperado 1 desde caché, obtenidos %d", len(list))
	}
}
