package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// Cache mock sencillo: para estos tests solo ocupamos simular cache miss.
type mockCache struct{}

func (mockCache) Get(ctx context.Context, key string, dest any) error {
	return errors.New("cache miss")
}
func (mockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error { return nil }
func (mockCache) Del(ctx context.Context, keys ...string) error                           { return nil }
func (mockCache) DelByPattern(ctx context.Context, pattern string) error                  { return nil }

// Estos repos son mocks en memoria. La idea es probar el handler sin depender de DB real.
type mockUserRepo struct{ users map[string]*domain.User }

func newMockUserRepo() *mockUserRepo { return &mockUserRepo{users: map[string]*domain.User{}} }
func (r *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		copy := *u
		return &copy, nil
	}
	return nil, domain.ErrNotFound
}
func (r *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			copy := *u
			return &copy, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	copy := *u
	r.users[u.ID] = &copy
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
	copy := *u
	return &copy, nil
}
func (r *mockUserRepo) Delete(ctx context.Context, id string) error { delete(r.users, id); return nil }

type mockRestaurantRepo struct{ restaurants map[string]*domain.Restaurant }

func newMockRestaurantRepo() *mockRestaurantRepo {
	return &mockRestaurantRepo{restaurants: map[string]*domain.Restaurant{}}
}
func (r *mockRestaurantRepo) Create(ctx context.Context, rest *domain.Restaurant) error {
	copy := *rest
	r.restaurants[rest.ID] = &copy
	return nil
}
func (r *mockRestaurantRepo) FindByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	if x, ok := r.restaurants[id]; ok {
		copy := *x
		return &copy, nil
	}
	return nil, domain.ErrNotFound
}
func (r *mockRestaurantRepo) FindAll(ctx context.Context) ([]domain.Restaurant, error) {
	out := []domain.Restaurant{}
	for _, v := range r.restaurants {
		out = append(out, *v)
	}
	return out, nil
}

type mockProductRepo struct{ products map[string]*domain.Product }

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{products: map[string]*domain.Product{}}
}
func (r *mockProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	if p, ok := r.products[id]; ok {
		copy := *p
		return &copy, nil
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
	copy := *p
	r.products[p.ID] = &copy
	return nil
}
func (r *mockProductRepo) Update(ctx context.Context, p *domain.Product) error {
	copy := *p
	r.products[p.ID] = &copy
	return nil
}
func (r *mockProductRepo) Delete(ctx context.Context, id string) error {
	delete(r.products, id)
	return nil
}

type mockMenuRepo struct{ menus map[string]*domain.Menu }

func newMockMenuRepo() *mockMenuRepo { return &mockMenuRepo{menus: map[string]*domain.Menu{}} }
func (r *mockMenuRepo) Create(ctx context.Context, m *domain.Menu) error {
	copy := *m
	r.menus[m.ID] = &copy
	return nil
}
func (r *mockMenuRepo) FindByID(ctx context.Context, id string) (*domain.Menu, error) {
	if m, ok := r.menus[id]; ok {
		copy := *m
		return &copy, nil
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
	copy := *m
	return &copy, nil
}
func (r *mockMenuRepo) Delete(ctx context.Context, id string) error { delete(r.menus, id); return nil }

type mockReservationRepo struct {
	reservations map[string]*domain.Reservation
	available    int
}

func newMockReservationRepo() *mockReservationRepo {
	return &mockReservationRepo{reservations: map[string]*domain.Reservation{}, available: 10}
}
func (r *mockReservationRepo) Create(ctx context.Context, res *domain.Reservation) error {
	copy := *res
	r.reservations[res.ID] = &copy
	return nil
}
func (r *mockReservationRepo) FindByID(ctx context.Context, id string) (*domain.Reservation, error) {
	if x, ok := r.reservations[id]; ok {
		copy := *x
		return &copy, nil
	}
	return nil, domain.ErrNotFound
}
func (r *mockReservationRepo) Cancel(ctx context.Context, id string) error {
	if x, ok := r.reservations[id]; ok {
		x.Status = domain.StatusCancelled
		return nil
	}
	return domain.ErrNotFound
}
func (r *mockReservationRepo) CheckAvailability(ctx context.Context, restaurantID string, partySize int) (int, error) {
	return r.available, nil
}

type mockOrderRepo struct{ orders map[string]*domain.Order }

func newMockOrderRepo() *mockOrderRepo { return &mockOrderRepo{orders: map[string]*domain.Order{}} }
func (r *mockOrderRepo) Create(ctx context.Context, o *domain.Order) error {
	copy := *o
	r.orders[o.ID] = &copy
	return nil
}
func (r *mockOrderRepo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	if o, ok := r.orders[id]; ok {
		copy := *o
		return &copy, nil
	}
	return nil, domain.ErrNotFound
}

// Lo dejo en test mode para que gin no meta ruido extra en la salida.
func setupGin() { gin.SetMode(gin.TestMode) }

// Helper chiquito para no repetir la misma creación de requests en cada test.
func performJSON(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func ginBindingJSON(data []byte, dest any) error { return json.Unmarshal(data, dest) }

// Si falla, dejo el body en el mensaje porque ayuda bastante a ver qué salió mal.
func requireStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Fatalf("status esperado %d, obtenido %d. body=%s", expected, w.Code, strings.TrimSpace(w.Body.String()))
	}
}

// Caso básico: registrar usuario y luego loguearlo con el mismo correo.
func TestAuthHandlerRegisterAndLogin(t *testing.T) {
	setupGin()
	users := newMockUserRepo()
	h := NewAuthHandler(service.NewAuthService(users, "secret-super-largo-123", time.Hour))
	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)

	registerBody := domain.RegisterRequest{Name: "Bea", Email: "bea@example.com", Password: "123456", Role: domain.RoleClient}
	w := performJSON(r, http.MethodPost, "/auth/register", registerBody)
	requireStatus(t, w, http.StatusCreated)
	if !strings.Contains(w.Body.String(), "token") {
		t.Fatalf("respuesta no contiene token: %s", w.Body.String())
	}

	loginBody := domain.LoginRequest{Email: "bea@example.com", Password: "123456"}
	w = performJSON(r, http.MethodPost, "/auth/login", loginBody)
	requireStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "bea@example.com") {
		t.Fatalf("respuesta no contiene usuario esperado: %s", w.Body.String())
	}
}
