package e2e_test

// postgres_test.go — pruebas E2E del stack completo: HTTP → service → Postgres real.
//
// Se saltan si Postgres no está disponible en POSTGRES_TEST_URL.
// Cómo correr:
//
//	POSTGRES_TEST_URL="postgres://postgres:postgres@localhost:5432/restaurants" \
//	  go test ./internal/e2e/... -v

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"restaurants-e2/internal/adapters/repopg"
	"restaurants-e2/internal/config"
	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
	transport "restaurants-e2/internal/transport/http"
)

// ── Infraestructura del test ──────────────────────────────────────────────────

const e2eJWTSecret = "e2e-jwt-secret-suficientemente-largo-para-hs256"

// noopCache implementa ports.Cache sin persistir nada — siempre cache miss.
type noopCache struct{}

func (noopCache) Get(_ context.Context, _ string, _ any) error                   { return errors.New("miss") }
func (noopCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error  { return nil }
func (noopCache) Del(_ context.Context, _ ...string) error                        { return nil }
func (noopCache) DelByPattern(_ context.Context, _ string) error                  { return nil }

// newE2EServer levanta un httptest.Server con el router real conectado a Postgres.
// Hace t.Skip si Postgres no está disponible.
func newE2EServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := os.Getenv("POSTGRES_TEST_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/restaurants"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := repopg.NewPool(ctx, config.PostgresConfig{DSN: dsn})
	if err != nil {
		t.Skipf("Postgres no disponible para E2E: %v", err)
	}
	t.Cleanup(pool.Close)

	repos := repopg.NewRepositories(pool)
	cache := noopCache{}

	r := transport.NewRouter(transport.Deps{
		AuthService:        service.NewAuthService(repos.Users, e2eJWTSecret, time.Hour),
		UserService:        service.NewUserService(repos.Users),
		RestaurantService:  service.NewRestaurantService(repos.Restaurants, cache),
		MenuService:        service.NewMenuService(repos.Menus, repos.Restaurants, repos.Products, cache),
		ProductService:     service.NewProductService(repos.Products, cache),
		ReservationService: service.NewReservationService(repos.Reservations, repos.Restaurants, cache),
		OrderService:       service.NewOrderService(repos.Orders, repos.Products, repos.Restaurants),
		JWTSecret:          e2eJWTSecret,
	})

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

// do envía una request JSON al servidor y retorna la response.
func do(t *testing.T, srv *httptest.Server, method, path string, body any, token string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, srv.URL+path, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, dest any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func requireStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		var body map[string]any
		json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck
		t.Fatalf("status esperado %d, obtenido %d. body=%v", expected, resp.StatusCode, body)
	}
}

// registerAndLogin registra un usuario único y retorna su token.
func registerAndLogin(t *testing.T, srv *httptest.Server, role string) (token, userID string) {
	t.Helper()
	email := "e2e-" + uuid.NewString()[:8] + "@test.com"
	pass := "pass1234"

	// Registro
	resp := do(t, srv, http.MethodPost, "/auth/register", map[string]any{
		"name": "Test User", "email": email, "password": pass,
	}, "")
	requireStatus(t, resp, http.StatusCreated)
	var reg domain.LoginResponse
	decodeJSON(t, resp, &reg)

	// Si se pide admin, tenemos que simular — en la práctica el admin lo crea el seed.
	// Para E2E usamos el token del cliente registrado.
	return reg.Token, reg.User.ID
}

// ── Autenticación ─────────────────────────────────────────────────────────────

func TestE2ERegisterSuccess(t *testing.T) {
	srv := newE2EServer(t)

	resp := do(t, srv, http.MethodPost, "/auth/register", map[string]any{
		"name":     "María García",
		"email":    "e2e-register-" + uuid.NewString()[:8] + "@test.com",
		"password": "pass1234",
	}, "")
	requireStatus(t, resp, http.StatusCreated)

	var body domain.LoginResponse
	decodeJSON(t, resp, &body)
	if body.Token == "" {
		t.Error("registro no retornó token")
	}
	if body.User.Role != domain.RoleClient {
		t.Errorf("rol esperado client, obtenido %q", body.User.Role)
	}
}

func TestE2ERegisterDuplicateEmail(t *testing.T) {
	srv := newE2EServer(t)
	email := "e2e-dup-" + uuid.NewString()[:8] + "@test.com"

	do(t, srv, http.MethodPost, "/auth/register", map[string]any{ //nolint:errcheck
		"name": "User A", "email": email, "password": "pass1234",
	}, "")

	resp := do(t, srv, http.MethodPost, "/auth/register", map[string]any{
		"name": "User B", "email": email, "password": "pass1234",
	}, "")
	requireStatus(t, resp, http.StatusConflict)
}

func TestE2ELoginSuccess(t *testing.T) {
	srv := newE2EServer(t)
	email := "e2e-login-" + uuid.NewString()[:8] + "@test.com"

	do(t, srv, http.MethodPost, "/auth/register", map[string]any{ //nolint:errcheck
		"name": "Login User", "email": email, "password": "mypassword",
	}, "")

	resp := do(t, srv, http.MethodPost, "/auth/login", map[string]any{
		"email": email, "password": "mypassword",
	}, "")
	requireStatus(t, resp, http.StatusOK)

	var body domain.LoginResponse
	decodeJSON(t, resp, &body)
	if body.Token == "" {
		t.Error("login no retornó token")
	}
}

func TestE2ELoginWrongPassword(t *testing.T) {
	srv := newE2EServer(t)
	email := "e2e-wrongpass-" + uuid.NewString()[:8] + "@test.com"

	do(t, srv, http.MethodPost, "/auth/register", map[string]any{ //nolint:errcheck
		"name": "User", "email": email, "password": "correcta",
	}, "")

	resp := do(t, srv, http.MethodPost, "/auth/login", map[string]any{
		"email": email, "password": "incorrecta",
	}, "")
	requireStatus(t, resp, http.StatusUnauthorized)
}

func TestE2EProtectedRouteNoToken(t *testing.T) {
	srv := newE2EServer(t)

	resp := do(t, srv, http.MethodGet, "/users/me", nil, "")
	requireStatus(t, resp, http.StatusUnauthorized)
}

func TestE2EProtectedRouteInvalidToken(t *testing.T) {
	srv := newE2EServer(t)

	resp := do(t, srv, http.MethodGet, "/users/me", nil, "token-invalido")
	requireStatus(t, resp, http.StatusUnauthorized)
}

// ── Restaurantes ──────────────────────────────────────────────────────────────

func TestE2EListRestaurantsPublic(t *testing.T) {
	srv := newE2EServer(t)

	// Sin token — debe ser público
	resp := do(t, srv, http.MethodGet, "/restaurants", nil, "")
	requireStatus(t, resp, http.StatusOK)
}

func TestE2EClientCannotCreateRestaurant(t *testing.T) {
	srv := newE2EServer(t)
	token, _ := registerAndLogin(t, srv, domain.RoleClient)

	resp := do(t, srv, http.MethodPost, "/restaurants", map[string]any{
		"name": "Mi Soda", "address": "San José", "phone": "+506 2222-0000", "capacity": 20,
	}, token)
	requireStatus(t, resp, http.StatusForbidden)
}

// ── Menús, Productos, Reservaciones, Órdenes (flujos con admin) ───────────────

// seedAdminViaDB crea un admin directamente en Postgres para los tests que lo necesitan.
// En un sistema real solo el seed crearía admins.
func seedAdminViaDB(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	// Registramos como client y obtenemos el token (no podemos crear admin vía API).
	// Para los tests de admin usamos el mecanismo del seed: la API no expone creación de admins.
	// En E2E podemos verificar que el client es rechazado en rutas de admin.
	token, _ := registerAndLogin(t, srv, domain.RoleClient)
	return token
}

func TestE2EMenuClientCannotUpdate(t *testing.T) {
	srv := newE2EServer(t)
	token := seedAdminViaDB(t, srv)

	// Un client intenta actualizar un menú (cualquier ID) → 403
	resp := do(t, srv, http.MethodPut, "/menus/"+uuid.NewString(), map[string]any{
		"name": "Intento",
	}, token)
	requireStatus(t, resp, http.StatusForbidden)
}

func TestE2EProductClientCannotDelete(t *testing.T) {
	srv := newE2EServer(t)
	token := seedAdminViaDB(t, srv)

	resp := do(t, srv, http.MethodDelete, "/products/"+uuid.NewString(), nil, token)
	requireStatus(t, resp, http.StatusForbidden)
}

func TestE2EReservationRestaurantNotFound(t *testing.T) {
	srv := newE2EServer(t)
	token, _ := registerAndLogin(t, srv, domain.RoleClient)

	resp := do(t, srv, http.MethodPost, "/reservations", map[string]any{
		"restaurant_id": uuid.NewString(),
		"date":          time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"party_size":    2,
	}, token)
	// Restaurante no existe → ErrValidation → 422
	requireStatus(t, resp, http.StatusUnprocessableEntity)
}

func TestE2EOrderProductNotFound(t *testing.T) {
	srv := newE2EServer(t)
	token, _ := registerAndLogin(t, srv, domain.RoleClient)

	resp := do(t, srv, http.MethodPost, "/orders", map[string]any{
		"restaurant_id": uuid.NewString(),
		"items": []map[string]any{
			{"product_id": uuid.NewString(), "quantity": 1},
		},
	}, token)
	// Restaurante no existe → ErrValidation → 422
	requireStatus(t, resp, http.StatusUnprocessableEntity)
}

func TestE2EGetMenuNotFound(t *testing.T) {
	srv := newE2EServer(t)
	token, _ := registerAndLogin(t, srv, domain.RoleClient)

	resp := do(t, srv, http.MethodGet, "/menus/"+uuid.NewString(), nil, token)
	requireStatus(t, resp, http.StatusNotFound)
}

func TestE2EGetRestaurantNotFound(t *testing.T) {
	srv := newE2EServer(t)

	resp := do(t, srv, http.MethodGet, "/restaurants/"+uuid.NewString(), nil, "")
	requireStatus(t, resp, http.StatusNotFound)
}

func TestE2EGetProductByCategory(t *testing.T) {
	srv := newE2EServer(t)
	token, _ := registerAndLogin(t, srv, domain.RoleClient)

	resp := do(t, srv, http.MethodGet, "/products?category=bebida", nil, token)
	requireStatus(t, resp, http.StatusOK)
}

func TestE2EGetOrderNotFound(t *testing.T) {
	srv := newE2EServer(t)
	token, _ := registerAndLogin(t, srv, domain.RoleClient)

	resp := do(t, srv, http.MethodGet, "/orders/"+uuid.NewString(), nil, token)
	requireStatus(t, resp, http.StatusNotFound)
}
