package http

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"restaurants-e2/internal/auth"
	"restaurants-e2/internal/domain"
)

// RequestID genera un UUID por request y lo inyecta en el contexto de Gin
// y en el header X-Request-ID de la respuesta. Facilita correlacionar logs
// con errores reportados por el cliente.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New().String()
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// AuthMiddleware valida el JWT del header Authorization: Bearer <token>.
// Si es válido, inyecta user_id, role y email en el contexto de Gin
// para que los handlers los consuman con c.GetString("user_id").
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "token requerido"})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.Parse(token, secret)
		if err != nil {
			renderError(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// AdminOnly es un middleware que debe usarse después de AuthMiddleware.
// Rechaza con 403 si el usuario autenticado no tiene rol "admin".
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != domain.RoleAdmin {
			renderError(c, domain.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
