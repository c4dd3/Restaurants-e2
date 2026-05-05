package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/auth"
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
func (r *mockMenuRepo) Delete(ctx context.Context, id string) error {
	if _, ok := r.menus[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.menus, id)
	return nil
}

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

	registerBody := domain.RegisterRequest{Name: "Bea", Email: "bea@example.com", Password: "123456"}
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

// Reviso el traductor de errores HTTP, incluyendo los errores envueltos con %w.
func TestRenderErrorStatuses(t *testing.T) {
	setupGin()

	cases := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{"not found", domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{"invalid credentials", domain.ErrInvalidCredentials, http.StatusUnauthorized, "invalid_credentials"},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"conflict", fmt.Errorf("correo repetido: %w", domain.ErrConflict), http.StatusConflict, "conflict"},
		{"validation", fmt.Errorf("nombre vacío: %w", domain.ErrValidation), http.StatusUnprocessableEntity, "validation_error"},
		{"internal", errors.New("db apagada"), http.StatusInternalServerError, "internal_server_error"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/err", func(c *gin.Context) { renderError(c, tc.err) })

			w := performJSON(r, http.MethodGet, "/err", nil)
			requireStatus(t, w, tc.status)
			if !strings.Contains(w.Body.String(), tc.body) {
				t.Fatalf("respuesta inesperada: %s", w.Body.String())
			}
		})
	}
}

// RequestID es pequeño, pero es parte del router real y conviene cubrirlo.
func TestRequestIDMiddleware(t *testing.T) {
	setupGin()
	r := gin.New()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) {
		if c.GetString("request_id") == "" {
			t.Fatal("no guardó request_id en el contexto")
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := performJSON(r, http.MethodGet, "/ping", nil)
	requireStatus(t, w, http.StatusOK)
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("no escribió X-Request-ID")
	}
}

func TestAuthMiddleware(t *testing.T) {
	setupGin()
	secret := "secret-super-largo-123"

	makeRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(AuthMiddleware(secret))
		r.GET("/private", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id": c.GetString("user_id"),
				"email":   c.GetString("email"),
				"role":    c.GetString("role"),
			})
		})
		return r
	}

	t.Run("sin token", func(t *testing.T) {
		w := performJSON(makeRouter(), http.MethodGet, "/private", nil)
		requireStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("token inválido", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		req.Header.Set("Authorization", "Bearer token-malo")
		w := httptest.NewRecorder()
		makeRouter().ServeHTTP(w, req)
		requireStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("token válido", func(t *testing.T) {
		token, err := auth.Sign("user-1", "bea@example.com", domain.RoleAdmin, secret, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		makeRouter().ServeHTTP(w, req)
		requireStatus(t, w, http.StatusOK)
		if !strings.Contains(w.Body.String(), "bea@example.com") || !strings.Contains(w.Body.String(), domain.RoleAdmin) {
			t.Fatalf("claims no llegaron al handler: %s", w.Body.String())
		}
	})
}

func TestAdminOnlyMiddleware(t *testing.T) {
	setupGin()

	makeRouter := func(role string) *gin.Engine {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("role", role)
			c.Next()
		})
		r.Use(AdminOnly())
		r.GET("/admin", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
		return r
	}

	w := performJSON(makeRouter(domain.RoleClient), http.MethodGet, "/admin", nil)
	requireStatus(t, w, http.StatusForbidden)

	w = performJSON(makeRouter(domain.RoleAdmin), http.MethodGet, "/admin", nil)
	requireStatus(t, w, http.StatusOK)
}

type mockSearchIndex struct {
	items []domain.Product
	err   error
}

func (m *mockSearchIndex) IndexProduct(ctx context.Context, p *domain.Product) error { return m.err }
func (m *mockSearchIndex) BulkIndexProducts(ctx context.Context, ps []domain.Product) error {
	return m.err
}
func (m *mockSearchIndex) SearchProducts(ctx context.Context, query string, limit int) ([]domain.Product, error) {
	return m.items, m.err
}
func (m *mockSearchIndex) SearchByCategory(ctx context.Context, category string, limit int) ([]domain.Product, error) {
	return m.items, m.err
}
func (m *mockSearchIndex) DeleteProduct(ctx context.Context, id string) error { return m.err }

func TestSearchHandlerRoutes(t *testing.T) {
	setupGin()
	products := newMockProductRepo()
	_ = products.Create(context.Background(), &domain.Product{ID: "prod-1", Name: "Pizza", Category: "pizzas"})

	h := NewSearchHandler(&mockSearchIndex{items: []domain.Product{{ID: "prod-1", Name: "Pizza", Category: "pizzas"}}}, products)
	r := gin.New()
	h.RegisterRoutes(r)

	w := performJSON(r, http.MethodGet, "/search/products", nil)
	requireStatus(t, w, http.StatusBadRequest)

	w = performJSON(r, http.MethodGet, "/search/products?q=pizza&limit=2", nil)
	requireStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "Pizza") {
		t.Fatalf("no devolvió producto esperado: %s", w.Body.String())
	}

	w = performJSON(r, http.MethodGet, "/search/products/category/pizzas?limit=99", nil)
	requireStatus(t, w, http.StatusOK)

	w = performJSON(r, http.MethodPost, "/search/reindex", nil)
	requireStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "indexed") {
		t.Fatalf("respuesta de reindex inesperada: %s", w.Body.String())
	}
}

func TestSearchHandlerErrors(t *testing.T) {
	setupGin()
	h := NewSearchHandler(&mockSearchIndex{err: errors.New("elastic down")}, newMockProductRepo())
	r := gin.New()
	h.RegisterRoutes(r)

	w := performJSON(r, http.MethodGet, "/search/products?q=pizza", nil)
	requireStatus(t, w, http.StatusServiceUnavailable)

	w = performJSON(r, http.MethodGet, "/search/products/category/pizzas", nil)
	requireStatus(t, w, http.StatusServiceUnavailable)

	products := newMockProductRepo()
	_ = products.Create(context.Background(), &domain.Product{ID: "prod-1", Name: "Pizza"})
	h = NewSearchHandler(&mockSearchIndex{err: errors.New("bulk fail")}, products)
	r = gin.New()
	h.RegisterRoutes(r)
	w = performJSON(r, http.MethodPost, "/search/reindex", nil)
	requireStatus(t, w, http.StatusInternalServerError)
}

func TestParseLimit(t *testing.T) {
	cases := map[string]int{
		"":    20,
		"abc": 20,
		"-1":  20,
		"0":   20,
		"10":  10,
		"200": 50,
	}
	for raw, expected := range cases {
		if got := parseLimit(raw); got != expected {
			t.Fatalf("parseLimit(%q)=%d, esperado %d", raw, got, expected)
		}
	}
}

func TestNewRouterHealthAndProtectedRoute(t *testing.T) {
	setupGin()
	users := newMockUserRepo()
	rests := newMockRestaurantRepo()
	menus := newMockMenuRepo()
	products := newMockProductRepo()
	reservations := newMockReservationRepo()
	orders := newMockOrderRepo()
	cache := mockCache{}

	r := NewRouter(Deps{
		AuthService:        service.NewAuthService(users, "secret-super-largo-123", time.Hour),
		UserService:        service.NewUserService(users),
		RestaurantService:  service.NewRestaurantService(rests, cache),
		MenuService:        service.NewMenuService(menus, rests, products, cache),
		ProductService:     service.NewProductService(products, cache),
		ReservationService: service.NewReservationService(reservations, rests, cache),
		OrderService:       service.NewOrderService(orders, products, rests),
		JWTSecret:          "secret-super-largo-123",
	})

	w := performJSON(r, http.MethodGet, "/health", nil)
	requireStatus(t, w, http.StatusOK)

	// Sin JWT debe quedarse en middleware, con eso cubrimos parte del router real.
	w = performJSON(r, http.MethodGet, "/users/me", nil)
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestAuthHandlerBadRequestsAndInvalidLogin(t *testing.T) {
	setupGin()
	users := newMockUserRepo()
	h := NewAuthHandler(service.NewAuthService(users, "secret-super-largo-123", time.Hour))

	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)

	// JSON inválido: cubre la rama de bad request del register.
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	requireStatus(t, w, http.StatusBadRequest)

	// Register incompleto: también debe quedarse en validación/binding.
	w = performJSON(r, http.MethodPost, "/auth/register", map[string]any{
		"email": "bea@example.com",
	})
	requireStatus(t, w, http.StatusBadRequest)

	// Login inválido por formato de request.
	w = performJSON(r, http.MethodPost, "/auth/login", map[string]any{
		"email": "bea@example.com",
	})
	requireStatus(t, w, http.StatusBadRequest)

	// Login con usuario inexistente: cubre error de credenciales.
	w = performJSON(r, http.MethodPost, "/auth/login", domain.LoginRequest{
		Email:    "nadie@example.com",
		Password: "123456",
	})
	requireStatus(t, w, http.StatusUnauthorized)
}
