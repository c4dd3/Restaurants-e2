package service

// Los errores que los servicios retornan están definidos en internal/domain/errors.go
// porque los adapters también los necesitan. Este archivo solo documenta el mapeo a HTTP:
//
//   domain.ErrNotFound           → 404 Not Found
//   domain.ErrConflict           → 409 Conflict
//   domain.ErrInvalidCredentials → 401 Unauthorized
//   domain.ErrUnauthorized       → 401 Unauthorized
//   domain.ErrForbidden          → 403 Forbidden
//   domain.ErrValidation         → 422 Unprocessable Entity
//
// La traducción real vive en internal/transport/http/errors.go.
