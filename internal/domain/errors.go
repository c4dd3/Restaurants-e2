package domain

import "errors"

// Errores de dominio compartidos por adapters, services y transport.
// Los adapters los emiten; los services los propagan o generan nuevos;
// transport/http/errors.go los traduce a HTTP status codes:
//
//   ErrNotFound           → 404
//   ErrConflict           → 409
//   ErrInvalidCredentials → 401
//   ErrUnauthorized       → 401
//   ErrForbidden          → 403
//   ErrValidation         → 422
var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrValidation         = errors.New("validation error")
)
