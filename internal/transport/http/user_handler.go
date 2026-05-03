package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// UserHandler maneja operaciones sobre el perfil de usuarios autenticados.
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetMe godoc
// GET /users/me
// Requiere: JWT
// Response 200: domain.User (sin password)
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.svc.GetMe(c.Request.Context(), userID)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update godoc
// PUT /users/:id
// Requiere: JWT
// Body: { name?, email? }
// Reglas: admin puede editar cualquier usuario; client solo el suyo.
// Response 200: domain.User actualizado
func (h *UserHandler) Update(c *gin.Context) {
	callerID := c.GetString("user_id")
	callerRole := c.GetString("role")
	targetID := c.Param("id")

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	user, err := h.svc.Update(c.Request.Context(), callerID, callerRole, targetID, req)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete godoc
// DELETE /users/:id
// Requiere: JWT
// Reglas: admin puede eliminar cualquier usuario; client solo el suyo.
// Response 204: No Content
func (h *UserHandler) Delete(c *gin.Context) {
	callerID := c.GetString("user_id")
	callerRole := c.GetString("role")
	targetID := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), callerID, callerRole, targetID); err != nil {
		renderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
