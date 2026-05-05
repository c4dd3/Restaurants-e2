package service

// mocks_test.go — repositorios e interfaces en memoria para tests unitarios.
// Al vivir en `package service` (mismo paquete que el código), todos los
// archivos *_test.go del paquete pueden usarlos sin importar nada extra.
//
// Los mocks soportan inyección de errores mediante campos opcionales
// (findByIDErr, createErr, etc.). Cuando un campo es no-nil, el método
// devuelve ese error en lugar de ejecutar la lógica normal. Esto permite
// cubrir las ramas de propagación de errores en el service layer.

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"restaurants-e2/internal/domain"
)

// ── mockUserRepo ──────────────────────────────────────────────────────────────

type mockUserRepo struct {
	users          map[string]*domain.User
	findByEmailErr error // si no nil, FindByEmail devuelve este error
	createErr      error // si no nil, Create devuelve este error (antes del check de duplicados)
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: map[string]*domain.User{}}
}

func (r *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		c := *u
		return &c, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if r.findByEmailErr != nil {
		return nil, r.findByEmailErr
	}
	for _, u := range r.users {
		if u.Email == email {
			c := *u
			return &c, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	// Simula unique constraint en email
	for _, existing := range r.users {
		if existing.Email == u.Email {
			return domain.ErrConflict
		}
	}
	c := *u
	r.users[u.ID] = &c
	return nil
}

func (r *mockUserRepo) Update(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	c := *u
	return &c, nil
}

func (r *mockUserRepo) Delete(ctx context.Context, id string) error {
	delete(r.users, id)
	return nil
}

// ── mockRestaurantRepo ────────────────────────────────────────────────────────

type mockRestaurantRepo struct {
	restaurants map[string]*domain.Restaurant
	findByIDErr error // si no nil, FindByID devuelve este error
	findAllErr  error // si no nil, FindAll devuelve este error
	createErr   error // si no nil, Create devuelve este error
}

func newMockRestaurantRepo() *mockRestaurantRepo {
	return &mockRestaurantRepo{restaurants: map[string]*domain.Restaurant{}}
}

func (r *mockRestaurantRepo) Create(ctx context.Context, rest *domain.Restaurant) error {
	if r.createErr != nil {
		return r.createErr
	}
	c := *rest
	r.restaurants[rest.ID] = &c
	return nil
}

func (r *mockRestaurantRepo) FindByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	if r.findByIDErr != nil {
		return nil, r.findByIDErr
	}
	if x, ok := r.restaurants[id]; ok {
		c := *x
		return &c, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockRestaurantRepo) FindAll(ctx context.Context) ([]domain.Restaurant, error) {
	if r.findAllErr != nil {
		return nil, r.findAllErr
	}
	out := make([]domain.Restaurant, 0, len(r.restaurants))
	for _, v := range r.restaurants {
		out = append(out, *v)
	}
	return out, nil
}

// ── mockMenuRepo ──────────────────────────────────────────────────────────────

type mockMenuRepo struct {
	menus    map[string]*domain.Menu
	createErr error // si no nil, Create devuelve este error
}

func newMockMenuRepo() *mockMenuRepo {
	return &mockMenuRepo{menus: map[string]*domain.Menu{}}
}

func (r *mockMenuRepo) Create(ctx context.Context, m *domain.Menu) error {
	if r.createErr != nil {
		return r.createErr
	}
	c := *m
	r.menus[m.ID] = &c
	return nil
}

func (r *mockMenuRepo) FindByID(ctx context.Context, id string) (*domain.Menu, error) {
	if m, ok := r.menus[id]; ok {
		c := *m
		return &c, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockMenuRepo) Update(ctx context.Context, id string, req *domain.UpdateMenuRequest) (*domain.Menu, error) {
	m, ok := r.menus[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	c := *m
	return &c, nil
}

func (r *mockMenuRepo) Delete(ctx context.Context, id string) error {
	if _, ok := r.menus[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.menus, id)
	return nil
}

// ── mockProductRepo ───────────────────────────────────────────────────────────

type mockProductRepo struct {
	products        map[string]*domain.Product
	findByCategoryErr error // si no nil, FindByCategory devuelve este error
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{products: map[string]*domain.Product{}}
}

func (r *mockProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	if p, ok := r.products[id]; ok {
		c := *p
		return &c, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockProductRepo) FindByIDs(ctx context.Context, ids []string) ([]domain.Product, error) {
	out := []domain.Product{}
	for _, id := range ids {
		if p, ok := r.products[id]; ok {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (r *mockProductRepo) FindByCategory(ctx context.Context, category string) ([]domain.Product, error) {
	if r.findByCategoryErr != nil {
		return nil, r.findByCategoryErr
	}
	out := []domain.Product{}
	for _, p := range r.products {
		if p.Category == category {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (r *mockProductRepo) FindAll(ctx context.Context) ([]domain.Product, error) {
	out := []domain.Product{}
	for _, p := range r.products {
		out = append(out, *p)
	}
	return out, nil
}

func (r *mockProductRepo) Create(ctx context.Context, p *domain.Product) error {
	c := *p
	r.products[p.ID] = &c
	return nil
}

// Update retorna ErrNotFound si el producto no existe en el mapa.
// Esto es consistente con el comportamiento del repo real (Postgres).
func (r *mockProductRepo) Update(ctx context.Context, p *domain.Product) error {
	if _, ok := r.products[p.ID]; !ok {
		return domain.ErrNotFound
	}
	c := *p
	r.products[p.ID] = &c
	return nil
}

// Delete retorna ErrNotFound si el producto no existe en el mapa.
func (r *mockProductRepo) Delete(ctx context.Context, id string) error {
	if _, ok := r.products[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.products, id)
	return nil
}

// ── mockReservationRepo ───────────────────────────────────────────────────────

type mockReservationRepo struct {
	reservations  map[string]*domain.Reservation
	available     int
	checkAvailErr error // si no nil, CheckAvailability devuelve este error
	cancelErr     error // si no nil, Cancel devuelve este error (incluso si existe)
}

func newMockReservationRepo(available int) *mockReservationRepo {
	return &mockReservationRepo{reservations: map[string]*domain.Reservation{}, available: available}
}

func (r *mockReservationRepo) Create(ctx context.Context, res *domain.Reservation) error {
	c := *res
	r.reservations[res.ID] = &c
	return nil
}

func (r *mockReservationRepo) FindByID(ctx context.Context, id string) (*domain.Reservation, error) {
	if x, ok := r.reservations[id]; ok {
		c := *x
		return &c, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockReservationRepo) Cancel(ctx context.Context, id string) error {
	if r.cancelErr != nil {
		return r.cancelErr
	}
	if x, ok := r.reservations[id]; ok {
		x.Status = domain.StatusCancelled
		return nil
	}
	return domain.ErrNotFound
}

func (r *mockReservationRepo) CheckAvailability(ctx context.Context, restaurantID string, partySize int) (int, error) {
	if r.checkAvailErr != nil {
		return 0, r.checkAvailErr
	}
	return r.available, nil
}

// ── mockOrderRepo ─────────────────────────────────────────────────────────────

type mockOrderRepo struct {
	orders map[string]*domain.Order
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: map[string]*domain.Order{}}
}

func (r *mockOrderRepo) Create(ctx context.Context, o *domain.Order) error {
	c := *o
	r.orders[o.ID] = &c
	return nil
}

func (r *mockOrderRepo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	if o, ok := r.orders[id]; ok {
		c := *o
		return &c, nil
	}
	return nil, domain.ErrNotFound
}

// ── mockCache ─────────────────────────────────────────────────────────────────
// Soporta hits reales: Set guarda en store, Get desserializa via JSON.
// Crea con newMockCache() para empezar con store vacío (siempre miss).

type mockCache struct {
	store       map[string]any
	deletedKeys []string
}

func newMockCache() *mockCache {
	return &mockCache{store: map[string]any{}}
}

func (c *mockCache) Get(ctx context.Context, key string, dest any) error {
	v, ok := c.store[key]
	if !ok {
		return errors.New("cache miss")
	}
	// JSON round-trip para respetar los tipos del dest igual que Redis.
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *mockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.store[key] = value
	return nil
}

func (c *mockCache) Del(ctx context.Context, keys ...string) error {
	c.deletedKeys = append(c.deletedKeys, keys...)
	return nil
}

func (c *mockCache) DelByPattern(ctx context.Context, pattern string) error {
	c.deletedKeys = append(c.deletedKeys, pattern)
	return nil
}
