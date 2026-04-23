package service

// Este archivo centraliza los ERRORES DE DOMINIO que los services exponen.
//
// Filosofía:
//   - Los adapters devuelven errores crudos (pgx.ErrNoRows, mongo.ErrNoDocuments, ...).
//   - El service los normaliza a uno de estos errores de dominio.
//   - La capa transport/http traduce cada uno a un HTTP status code (ver transport/http/errors.go).
//
// De esta forma:
//   - El service no sabe nada de HTTP.
//   - El handler no sabe nada de pgx/mongo.
//
// Lista mínima a implementar:
//
//   ErrNotFound            → 404  (recurso inexistente)
//   ErrInvalidCredentials  → 401  (login fallido)
//   ErrUnauthorized        → 401  (token inválido/vencido)
//   ErrForbidden           → 403  (autenticado pero sin permisos — ej. no-admin)
//   ErrConflict            → 409  (violación de unicidad, reserva solapada, etc.)
//   ErrValidation          → 422  (DTO válido pero viola reglas de negocio)
//
// Usar `errors.New("...")` para cada uno. Opcionalmente envolver con fmt.Errorf
// para agregar contexto sin perder la identidad (errors.Is funciona con la base).
