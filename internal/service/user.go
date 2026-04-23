package service

// UserService — casos de uso sobre el perfil del usuario (después del login).
//
// Diferencia con AuthService:
//   - AuthService   → Register, Login (emite tokens, maneja passwords).
//   - UserService   → "ya estás adentro" — leer y modificar tu propio perfil.
//
// Dependencias:
//   - ports.UserRepository
//
// Métodos públicos:
//
//   GetMe(ctx, userID string) (*domain.User, error)
//     1. UserRepository.FindByID(ctx, userID).
//     2. Si no existe → ErrNotFound.
//     3. Devolver *domain.User (el handler NO serializa el password hash;
//        el struct User debe tener json:"-" en el campo password).
//
//   UpdateMe(ctx, userID, UpdateUserRequest) (*domain.User, error)
//     1. Validar que al menos un campo venga no vacío.
//     2. UserRepository.Update(ctx, id, req).
//     3. Devolver el usuario actualizado.
//
// Posibles extensiones (FUERA del alcance mínimo de Etapa 2):
//   - ChangePassword: requiere contraseña vieja, aplica bcrypt, persiste.
//   - ListUsers (solo admin): con paginación.
