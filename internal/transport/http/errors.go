package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
)

// renderError traduce un error de dominio a una respuesta HTTP JSON.
// Es el único lugar del sistema donde vive esta lógica — ningún handler
// repite el switch. Usa errors.Is para respetar error wrapping (%w).
func renderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})

	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})

	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})

	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})

	case errors.Is(err, domain.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "conflict"})

	case errors.Is(err, domain.ErrValidation):
		// ErrValidation es seguro de exponer al cliente — son errores de reglas de negocio,
		// no stacktraces ni detalles de BD.
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":  "validation_error",
			"detail": err.Error(),
		})

	default:
		// Error inesperado: loguear completo internamente, responder genérico al cliente.
		log.Printf("[error] unhandled: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
	}
}
