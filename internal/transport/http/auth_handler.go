package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// AuthHandler maneja registro y login. No requiere JWT — son los endpoints
// que emiten el token por primera vez.
type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register godoc
// POST /auth/register
// Body: { name, email, password }
// Response 201: { token, user } — rol asignado automáticamente como "client"
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	user, token, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, domain.LoginResponse{Token: token, User: *user})
}

// Login godoc
// POST /auth/login
// Body: { email, password }
// Response 200: { token, user }
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	user, token, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.LoginResponse{Token: token, User: *user})
}
