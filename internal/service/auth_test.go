package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"restaurants-e2/internal/auth"
	"restaurants-e2/internal/domain"
)

const testSecret = "jwt-secret-suficientemente-largo-para-tests"

func newAuthSvc(users *mockUserRepo) *AuthService {
	return NewAuthService(users, testSecret, time.Hour)
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestAuthServiceRegisterHappyPath(t *testing.T) {
	svc := newAuthSvc(newMockUserRepo())

	user, token, err := svc.Register(context.Background(), domain.RegisterRequest{
		Name:     "Carlos Mora",
		Email:    "carlos@example.com",
		Password: "pass1234",
	})

	if err != nil {
		t.Fatalf("Register inesperado: %v", err)
	}
	if user == nil || user.ID == "" {
		t.Fatal("Register no devolvió un usuario con ID")
	}
	if token == "" {
		t.Fatal("Register no devolvió token")
	}
}

func TestAuthServiceRegisterRoleAlwaysClient(t *testing.T) {
	// Aunque el request no tenga campo Role, el servicio debe asignar "client".
	svc := newAuthSvc(newMockUserRepo())

	user, _, err := svc.Register(context.Background(), domain.RegisterRequest{
		Name:     "María López",
		Email:    "maria@example.com",
		Password: "pass1234",
	})

	if err != nil {
		t.Fatalf("Register inesperado: %v", err)
	}
	if user.Role != domain.RoleClient {
		t.Errorf("rol esperado %q, obtenido %q", domain.RoleClient, user.Role)
	}
}

func TestAuthServiceRegisterDuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := newAuthSvc(repo)

	req := domain.RegisterRequest{Name: "Ana", Email: "ana@example.com", Password: "pass1234"}
	if _, _, err := svc.Register(context.Background(), req); err != nil {
		t.Fatalf("primer registro falló: %v", err)
	}

	_, _, err := svc.Register(context.Background(), req)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("esperado ErrConflict, obtenido %v", err)
	}
}

func TestAuthServiceRegisterPasswordIsHashed(t *testing.T) {
	repo := newMockUserRepo()
	svc := newAuthSvc(repo)

	plainPass := "mi-password-secreta"
	user, _, _ := svc.Register(context.Background(), domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: plainPass,
	})

	// La contraseña almacenada nunca debe ser texto claro.
	if user.Password == plainPass {
		t.Fatal("contraseña almacenada en texto claro")
	}
	// Pero debe ser verificable.
	if err := auth.VerifyPassword(user.Password, plainPass); err != nil {
		t.Fatalf("contraseña hasheada no verificable: %v", err)
	}
}

func TestAuthServiceRegisterTokenCarriesCorrectClaims(t *testing.T) {
	svc := newAuthSvc(newMockUserRepo())

	user, token, _ := svc.Register(context.Background(), domain.RegisterRequest{
		Name:     "Luis Vargas",
		Email:    "luis@example.com",
		Password: "pass1234",
	})

	claims, err := auth.Parse(token, testSecret)
	if err != nil {
		t.Fatalf("token inválido: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("UserID en token %q ≠ user.ID %q", claims.UserID, user.ID)
	}
	if claims.Email != "luis@example.com" {
		t.Errorf("Email en token %q ≠ esperado", claims.Email)
	}
	if claims.Role != domain.RoleClient {
		t.Errorf("Role en token %q ≠ client", claims.Role)
	}
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestAuthServiceLoginHappyPath(t *testing.T) {
	repo := newMockUserRepo()
	svc := newAuthSvc(repo)

	// Primero registramos para que haya un usuario real en el repo.
	svc.Register(context.Background(), domain.RegisterRequest{ //nolint:errcheck
		Name: "Pedro", Email: "pedro@example.com", Password: "pass1234",
	})

	user, token, err := svc.Login(context.Background(), domain.LoginRequest{
		Email:    "pedro@example.com",
		Password: "pass1234",
	})

	if err != nil {
		t.Fatalf("Login inesperado: %v", err)
	}
	if user == nil || user.Email != "pedro@example.com" {
		t.Fatalf("usuario incorrecto: %#v", user)
	}
	if token == "" {
		t.Fatal("token vacío en Login")
	}
}

func TestAuthServiceLoginWrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := newAuthSvc(repo)

	svc.Register(context.Background(), domain.RegisterRequest{ //nolint:errcheck
		Name: "Sofía", Email: "sofia@example.com", Password: "correcta",
	})

	_, _, err := svc.Login(context.Background(), domain.LoginRequest{
		Email:    "sofia@example.com",
		Password: "incorrecta",
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("esperado ErrInvalidCredentials, obtenido %v", err)
	}
}

func TestAuthServiceLoginEmailNotFound(t *testing.T) {
	svc := newAuthSvc(newMockUserRepo())

	_, _, err := svc.Login(context.Background(), domain.LoginRequest{
		Email:    "nadie@example.com",
		Password: "cualquier",
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("esperado ErrInvalidCredentials, obtenido %v", err)
	}
}

// ── Rutas de error en Register ────────────────────────────────────────────────

func TestAuthServiceRegisterFindByEmailError(t *testing.T) {
	// FindByEmail devuelve un error inesperado (no ErrNotFound) → debe propagarse.
	repo := newMockUserRepo()
	repo.findByEmailErr = errors.New("bd caída")
	svc := newAuthSvc(repo)

	_, _, err := svc.Register(context.Background(), domain.RegisterRequest{
		Name: "X", Email: "x@x.com", Password: "pass1234",
	})
	if err == nil {
		t.Fatal("esperaba error de BD en FindByEmail, obtuvo nil")
	}
}

func TestAuthServiceRegisterCreateError(t *testing.T) {
	// Create falla con error no-ErrConflict → debe propagarse.
	repo := newMockUserRepo()
	repo.createErr = errors.New("error de inserción")
	svc := newAuthSvc(repo)

	_, _, err := svc.Register(context.Background(), domain.RegisterRequest{
		Name: "X", Email: "nuevo@x.com", Password: "pass1234",
	})
	if err == nil {
		t.Fatal("esperaba error de BD en Create, obtuvo nil")
	}
}

func TestAuthServiceLoginMalformedHash(t *testing.T) {
	// Inyectamos un usuario cuyo campo Password es texto plano (no un hash bcrypt).
	// Esto hace que VerifyPassword devuelva un error distinto a ErrMismatchedHashAndPassword,
	// cubriendo la rama else del switch en service/auth.go.
	repo := newMockUserRepo()
	repo.users["u-broken"] = &domain.User{
		ID:       "u-broken",
		Name:     "Broken",
		Email:    "broken@example.com",
		Password: "no-es-un-hash-bcrypt",
		Role:     domain.RoleClient,
	}
	svc := newAuthSvc(repo)

	_, _, err := svc.Login(context.Background(), domain.LoginRequest{
		Email:    "broken@example.com",
		Password: "cualquier",
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("esperado ErrInvalidCredentials con hash malformado, obtenido %v", err)
	}
}
