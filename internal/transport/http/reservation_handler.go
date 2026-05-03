package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// ReservationHandler maneja operaciones sobre reservas de mesa.
type ReservationHandler struct {
	svc *service.ReservationService
}

func NewReservationHandler(svc *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{svc: svc}
}

// Create godoc
// POST /reservations
// Requiere: JWT
// Body: { restaurant_id, date, party_size, notes? }
// Response 201: domain.Reservation
// Errores:
//   - 422 si el restaurante no existe o los datos son inválidos
//   - 409 si no hay disponibilidad en esa ventana de tiempo
func (h *ReservationHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req domain.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "detail": err.Error()})
		return
	}

	reservation, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

// Cancel godoc
// DELETE /reservations/:id
// Requiere: JWT
// Solo el dueño de la reserva puede cancelarla (el service lo verifica).
// Response 204: No Content
func (h *ReservationHandler) Cancel(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.svc.Cancel(c.Request.Context(), userID, id); err != nil {
		renderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
