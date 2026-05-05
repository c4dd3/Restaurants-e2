package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// Flujo del usuario autenticado: ver su perfil, actualizarlo y luego eliminarlo.
func TestUserHandlerGetMeUpdateDelete(t *testing.T) {
	setupGin()
	users := newMockUserRepo()
	_ = users.Create(nil, &domain.User{ID: "user-1", Name: "Bea", Email: "bea@example.com", Role: domain.RoleClient})
	h := NewUserHandler(service.NewUserService(users))
	r := gin.New()
	r.GET("/users/me", func(c *gin.Context) { c.Set("user_id", "user-1"); h.GetMe(c) })
	r.PUT("/users/:id", func(c *gin.Context) { c.Set("user_id", "user-1"); c.Set("role", domain.RoleClient); h.Update(c) })
	r.DELETE("/users/:id", func(c *gin.Context) { c.Set("user_id", "user-1"); c.Set("role", domain.RoleClient); h.Delete(c) })

	w := performJSON(r, http.MethodGet, "/users/me", nil)
	requireStatus(t, w, http.StatusOK)

	w = performJSON(r, http.MethodPut, "/users/user-1", domain.UpdateUserRequest{Name: "Beatriz"})
	requireStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "Beatriz") {
		t.Fatalf("usuario no actualizado: %s", w.Body.String())
	}

	w = performJSON(r, http.MethodDelete, "/users/user-1", nil)
	requireStatus(t, w, http.StatusNoContent)
}

func TestUserHandlerErrors(t *testing.T) {
	setupGin()
	users := newMockUserRepo()
	h := NewUserHandler(service.NewUserService(users))

	r := gin.New()
	r.GET("/users/me", func(c *gin.Context) {
		c.Set("user_id", "no-existe")
		h.GetMe(c)
	})
	r.PUT("/users/:id", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("role", domain.RoleClient)
		h.Update(c)
	})
	r.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("role", domain.RoleClient)
		h.Delete(c)
	})

	// Usuario actual no existe.
	w := performJSON(r, http.MethodGet, "/users/me", nil)
	requireStatus(t, w, http.StatusNotFound)

	// Un cliente no debería poder actualizar otro usuario.
	w = performJSON(r, http.MethodPut, "/users/otro-user", domain.UpdateUserRequest{Name: "Otro"})
	requireStatus(t, w, http.StatusForbidden)

	// JSON malo en update.
	req := httptest.NewRequest(http.MethodPut, "/users/user-1", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	requireStatus(t, w, http.StatusBadRequest)
}
