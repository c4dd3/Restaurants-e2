package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"restaurants-e2/internal/auth"
	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// AuthService gestiona registro y login. Es el único lugar del sistema
// que maneja contraseñas en claro — una vez hasheadas nadie más las ve.
type AuthService struct {
	users  ports.UserRepository
	secret string
	ttl    time.Duration
}

// NewAuthService construye el servicio inyectando sus dependencias.
func NewAuthService(users ports.UserRepository, secret string, ttl time.Duration) *AuthService {
	return &AuthService{users: users, secret: secret, ttl: ttl}
}

// Register crea un usuario nuevo y emite su JWT.
// Retorna domain.ErrConflict si el email ya existe.
func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, string, error) {
	// Verificar email único antes de hashear (operación barata primero).
	existing, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", domain.ErrConflict
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, "", err
	}

	// Usar el rol del request si es válido; en cualquier otro caso asignar "client".
	role := domain.RoleClient
	if req.Role == domain.RoleAdmin || req.Role == domain.RoleClient {
		role = req.Role
	}

	u := &domain.User{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: hash,
		Role:     role,
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", err
	}

	token, err := auth.Sign(u.ID, u.Email, u.Role, s.secret, s.ttl)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

// Login valida credenciales y emite un JWT.
// Siempre retorna domain.ErrInvalidCredentials ante cualquier fallo de autenticación
// (no se distingue "email inexistente" de "password errónea" — evita enumeración).
func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.User, string, error) {
	u, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	if err := auth.VerifyPassword(u.Password, req.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", domain.ErrInvalidCredentials
	}

	token, err := auth.Sign(u.ID, u.Email, u.Role, s.secret, s.ttl)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}
