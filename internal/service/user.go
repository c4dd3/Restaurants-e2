package service

import (
	"context"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// UserService gestiona el perfil de usuarios.
// Se diferencia de AuthService en que aquí el usuario ya está autenticado.
type UserService struct {
	users ports.UserRepository
}

// NewUserService construye el servicio inyectando sus dependencias.
func NewUserService(users ports.UserRepository) *UserService {
	return &UserService{users: users}
}

// GetMe devuelve el perfil del usuario autenticado.
// El campo Password no se serializa (json:"-"), es seguro devolver el struct directo.
func (s *UserService) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	return s.users.FindByID(ctx, userID)
}

// Update modifica name y/o email de un usuario.
// Reglas de autorización:
//   - Admin puede modificar cualquier usuario.
//   - Client solo puede modificar su propio perfil (callerID == targetID).
func (s *UserService) Update(ctx context.Context, callerID, callerRole, targetID string, req domain.UpdateUserRequest) (*domain.User, error) {
	if callerRole != domain.RoleAdmin && callerID != targetID {
		return nil, domain.ErrForbidden
	}
	if req.Name == "" && req.Email == "" {
		return nil, domain.ErrValidation
	}
	return s.users.Update(ctx, targetID, &req)
}

// Delete elimina un usuario por ID.
// Reglas de autorización:
//   - Admin puede eliminar cualquier usuario.
//   - Client solo puede eliminarse a sí mismo (callerID == targetID).
func (s *UserService) Delete(ctx context.Context, callerID, callerRole, targetID string) error {
	if callerRole != domain.RoleAdmin && callerID != targetID {
		return domain.ErrForbidden
	}
	return s.users.Delete(ctx, targetID)
}
