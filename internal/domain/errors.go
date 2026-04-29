package domain

import "errors"

var (
	// ErrNotFound se retorna cuando un recurso buscado no existe en la base de datos.
	ErrNotFound = errors.New("not found")

	// ErrConflict se retorna cuando se viola una restricción de unicidad (ej: email duplicado).
	ErrConflict = errors.New("conflict: resource already exists")

	// ErrForbidden se retorna cuando el usuario no tiene permisos para la operación.
	ErrForbidden = errors.New("forbidden")
)
