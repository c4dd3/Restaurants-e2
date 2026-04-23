package http

// errors.go — traductor de errores de dominio a respuestas HTTP.
//
// Función pública:
//
//   domainToHTTP(err error) (statusCode int, body gin.H)
//
// Regla general: cada error sentinel de service/errors.go mapea a un status
// específico. Errores NO reconocidos → 500 Internal Server Error y log.
//
// Tabla de mapeo:
//
//   err                              → status  → body
//   ─────────────────────────────────────────────────────────────────────────
//   service.ErrNotFound              → 404     → {"error": "not_found"}
//   service.ErrInvalidCredentials    → 401     → {"error": "invalid_credentials"}
//   service.ErrUnauthorized          → 401     → {"error": "unauthorized"}
//   service.ErrForbidden             → 403     → {"error": "forbidden"}
//   service.ErrConflict              → 409     → {"error": "conflict"}
//   service.ErrValidation            → 422     → {"error": "validation", "detail": err.Error()}
//   (desconocido)                    → 500     → {"error": "internal"}
//
// Implementación:
//
//   func domainToHTTP(err error) (int, gin.H) {
//       switch {
//       case errors.Is(err, service.ErrNotFound):
//           return 404, gin.H{"error": "not_found"}
//       case errors.Is(err, service.ErrInvalidCredentials):
//           return 401, gin.H{"error": "invalid_credentials"}
//       ...
//       default:
//           log.Printf("unhandled error: %v", err)
//           return 500, gin.H{"error": "internal"}
//       }
//   }
//
// Importante:
//   - Usar errors.Is para respetar error wrapping (fmt.Errorf("...: %w", err)).
//   - NO devolver err.Error() al cliente salvo en ErrValidation (los detalles
//     de validación son seguros y útiles; stacktraces/SQL errors NO).
//   - En 500, loguear el error completo pero responder genérico.
//
// Por qué NO usar panic/recover para errores de dominio:
//   - Panic se reserva para programming bugs (nil deref, etc.). Los errores
//     de negocio son valores normales y viajan por return. Gin Recovery()
//     solo protege contra panics inesperados → 500.
