package service

import (
	"context"
	"errors"
	"testing"

	"restaurants-e2/internal/domain"
)

func newUserSvc(users *mockUserRepo) *UserService {
	return NewUserService(users)
}

// seedUser agrega un usuario directamente en el mock para preparar el estado inicial.
func seedUser(repo *mockUserRepo, id, email, role string) {
	repo.users[id] = &domain.User{ID: id, Name: "Test User", Email: email, Role: role}
}

// ── GetMe ─────────────────────────────────────────────────────────────────────

func TestUserServiceGetMeFound(t *testing.T) {
	repo := newMockUserRepo()
	seedUser(repo, "u-1", "user@example.com", domain.RoleClient)
	svc := newUserSvc(repo)

	u, err := svc.GetMe(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("GetMe inesperado: %v", err)
	}
	if u.ID != "u-1" {
		t.Errorf("ID esperado u-1, obtenido %q", u.ID)
	}
}

func TestUserServiceGetMeNotFound(t *testing.T) {
	svc := newUserSvc(newMockUserRepo())

	_, err := svc.GetMe(context.Background(), "no-existe")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperado ErrNotFound, obtenido %v", err)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUserServiceUpdateOwnProfile(t *testing.T) {
	repo := newMockUserRepo()
	seedUser(repo, "u-1", "original@example.com", domain.RoleClient)
	svc := newUserSvc(repo)

	updated, err := svc.Update(context.Background(), "u-1", domain.RoleClient, "u-1",
		domain.UpdateUserRequest{Name: "Nuevo Nombre"})

	if err != nil {
		t.Fatalf("Update inesperado: %v", err)
	}
	if updated.Name != "Nuevo Nombre" {
		t.Errorf("nombre esperado 'Nuevo Nombre', obtenido %q", updated.Name)
	}
}

func TestUserServiceUpdateOtherForbidden(t *testing.T) {
	repo := newMockUserRepo()
	seedUser(repo, "u-1", "user1@example.com", domain.RoleClient)
	seedUser(repo, "u-2", "user2@example.com", domain.RoleClient)
	svc := newUserSvc(repo)

	_, err := svc.Update(context.Background(), "u-1", domain.RoleClient, "u-2",
		domain.UpdateUserRequest{Name: "Intento"})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

func TestUserServiceAdminCanUpdateAnyone(t *testing.T) {
	repo := newMockUserRepo()
	seedUser(repo, "admin-1", "admin@example.com", domain.RoleAdmin)
	seedUser(repo, "u-99", "user99@example.com", domain.RoleClient)
	svc := newUserSvc(repo)

	updated, err := svc.Update(context.Background(), "admin-1", domain.RoleAdmin, "u-99",
		domain.UpdateUserRequest{Email: "nuevo@example.com"})

	if err != nil {
		t.Fatalf("Admin Update inesperado: %v", err)
	}
	if updated.Email != "nuevo@example.com" {
		t.Errorf("email esperado 'nuevo@example.com', obtenido %q", updated.Email)
	}
}

func TestUserServiceUpdateEmptyFieldsValidation(t *testing.T) {
	repo := newMockUserRepo()
	seedUser(repo, "u-1", "user@example.com", domain.RoleClient)
	svc := newUserSvc(repo)

	// Ambos campos vacíos → ErrValidation
	_, err := svc.Update(context.Background(), "u-1", domain.RoleClient, "u-1",
		domain.UpdateUserRequest{})

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("esperado ErrValidation, obtenido %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestUserServiceDeleteSelf(t *testing.T) {
	repo := newMockUserRepo()
	seedUser(repo, "u-1", "user@example.com", domain.RoleClient)
	svc := newUserSvc(repo)

	if err := svc.Delete(context.Background(), "u-1", domain.RoleClient, "u-1"); err != nil {
		t.Fatalf("Delete inesperado: %v", err)
	}

	// Verificar que ya no existe en el repo
	if _, exists := repo.users["u-1"]; exists {
		t.Fatal("usuario sigue en el repo después de Delete")
	}
}

func TestUserServiceDeleteOtherForbidden(t *testing.T) {
	repo := newMockUserRepo()
	seedUser(repo, "u-1", "user1@example.com", domain.RoleClient)
	seedUser(repo, "u-2", "user2@example.com", domain.RoleClient)
	svc := newUserSvc(repo)

	err := svc.Delete(context.Background(), "u-1", domain.RoleClient, "u-2")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperado ErrForbidden, obtenido %v", err)
	}
}

func TestUserServiceAdminDeletesAnyone(t *testing.T) {
	repo := newMockUserRepo()
	seedUser(repo, "admin-1", "admin@example.com", domain.RoleAdmin)
	seedUser(repo, "u-5", "user5@example.com", domain.RoleClient)
	svc := newUserSvc(repo)

	if err := svc.Delete(context.Background(), "admin-1", domain.RoleAdmin, "u-5"); err != nil {
		t.Fatalf("Admin Delete inesperado: %v", err)
	}
	if _, exists := repo.users["u-5"]; exists {
		t.Fatal("usuario sigue en el repo después de Delete por admin")
	}
}
