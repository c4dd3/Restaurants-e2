package service

import (
	"context"
	"errors"
	"testing"

	"restaurants-e2/internal/domain"
)

func newRestaurantSvc(rests *mockRestaurantRepo, cache *mockCache) *RestaurantService {
	return NewRestaurantService(rests, cache)
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestRestaurantServiceCreateAdmin(t *testing.T) {
	repo := newMockRestaurantRepo()
	cache := newMockCache()
	svc := newRestaurantSvc(repo, cache)

	r, err := svc.Create(context.Background(), "admin-1", domain.RoleAdmin,
		domain.CreateRestaurantRequest{
			Name:     "Soda La Tica",
			Address:  "San José, Barrio Amón",
			Phone:    "+506 2222-3333",
			Capacity: 40,
		})

	if err != nil {
		t.Fatalf("Create inesperado: %v", err)
	}
	if r == nil || r.ID == "" {
		t.Fatal("restaurante sin ID")
	}
	if r.AdminID != "admin-1" {
		t.Errorf("AdminID esperado admin-1, obtenido %q", r.AdminID)
	}
	// La cache debe haber recibido una invalidación
	if len(cache.deletedKeys) == 0 {
		t.Error("Create no invalidó la cache")
	}
}

func TestRestaurantServiceCreateClientForbidden(t *testing.T) {
	svc := newRestaurantSvc(newMockRestaurantRepo(), newMockCache())

	_, err := svc.Create(context.Background(), "u-1", domain.RoleClient,
		domain.CreateRestaurantRequest{Name: "Mi Soda", Address: "X", Phone: "X", Capacity: 10})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestRestaurantServiceGetByIDCacheMiss(t *testing.T) {
	repo := newMockRestaurantRepo()
	// Insertar directo en el repo (saltando el servicio)
	repo.restaurants["rest-1"] = &domain.Restaurant{ID: "rest-1", Name: "El Fogón"}

	svc := newRestaurantSvc(repo, newMockCache())

	r, err := svc.GetByID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("GetByID inesperado: %v", err)
	}
	if r.Name != "El Fogón" {
		t.Errorf("nombre esperado 'El Fogón', obtenido %q", r.Name)
	}
}

func TestRestaurantServiceGetByIDNotFound(t *testing.T) {
	svc := newRestaurantSvc(newMockRestaurantRepo(), newMockCache())

	_, err := svc.GetByID(context.Background(), "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestRestaurantServiceList(t *testing.T) {
	repo := newMockRestaurantRepo()
	repo.restaurants["r-1"] = &domain.Restaurant{ID: "r-1", Name: "Restaurante A"}
	repo.restaurants["r-2"] = &domain.Restaurant{ID: "r-2", Name: "Restaurante B"}
	svc := newRestaurantSvc(repo, newMockCache())

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List inesperado: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("esperados 2 restaurantes, obtenidos %d", len(list))
	}
}

func TestRestaurantServiceListEmpty(t *testing.T) {
	svc := newRestaurantSvc(newMockRestaurantRepo(), newMockCache())

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List inesperado: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("esperada lista vacía, obtenidos %d elementos", len(list))
	}
}

// ── Cache HIT paths ───────────────────────────────────────────────────────────

func TestRestaurantServiceGetByIDCacheHit(t *testing.T) {
	cache := newMockCache()
	// Pre-popularar el cache — simula que ya se consultó antes.
	cached := &domain.Restaurant{ID: "rest-1", Name: "Desde Caché"}
	cache.Set(context.Background(), "restaurants:id:rest-1", cached, 0) //nolint:errcheck

	// El repo está vacío; si se consulta la BD el test falla.
	svc := newRestaurantSvc(newMockRestaurantRepo(), cache)

	r, err := svc.GetByID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("GetByID cache hit inesperado: %v", err)
	}
	if r.Name != "Desde Caché" {
		t.Errorf("esperado nombre 'Desde Caché', obtenido %q", r.Name)
	}
}

func TestRestaurantServiceListFindAllError(t *testing.T) {
	repo := newMockRestaurantRepo()
	repo.findAllErr = errors.New("bd caída")
	svc := newRestaurantSvc(repo, newMockCache())

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("esperaba error de BD en FindAll, obtuvo nil")
	}
}

func TestRestaurantServiceCreateError(t *testing.T) {
	repo := newMockRestaurantRepo()
	repo.createErr = errors.New("error al insertar")
	svc := newRestaurantSvc(repo, newMockCache())

	_, err := svc.Create(context.Background(), "admin-1", domain.RoleAdmin,
		domain.CreateRestaurantRequest{Name: "Test", Address: "X", Phone: "X", Capacity: 10})
	if err == nil {
		t.Fatal("esperaba error de BD en Create, obtuvo nil")
	}
}

func TestRestaurantServiceListCacheHit(t *testing.T) {
	cache := newMockCache()
	cachedList := []domain.Restaurant{
		{ID: "r-1", Name: "Del Caché A"},
		{ID: "r-2", Name: "Del Caché B"},
	}
	cache.Set(context.Background(), "restaurants:all", cachedList, 0) //nolint:errcheck

	svc := newRestaurantSvc(newMockRestaurantRepo(), cache)

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List cache hit inesperado: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("esperados 2 desde caché, obtenidos %d", len(list))
	}
}
