package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/service"
)

// Reserva feliz: crear una reserva válida y luego cancelarla.
func TestReservationHandlerCreateAndCancel(t *testing.T) {
	setupGin()
	rests := newMockRestaurantRepo()
	reservations := newMockReservationRepo()
	_ = rests.Create(nil, &domain.Restaurant{ID: "rest-1", Name: "Soda TEC", Capacity: 20})
	h := NewReservationHandler(service.NewReservationService(reservations, rests, mockCache{}))
	r := gin.New()
	r.POST("/reservations", func(c *gin.Context) { c.Set("user_id", "user-1"); h.Create(c) })
	r.DELETE("/reservations/:id", func(c *gin.Context) { c.Set("user_id", "user-1"); h.Cancel(c) })

	w := performJSON(r, http.MethodPost, "/reservations", domain.CreateReservationRequest{RestaurantID: "rest-1", Date: time.Now().Add(24 * time.Hour), PartySize: 2})
	requireStatus(t, w, http.StatusCreated)

	var created domain.Reservation
	if err := ginBindingJSON(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	w = performJSON(r, http.MethodDelete, "/reservations/"+created.ID, nil)
	requireStatus(t, w, http.StatusNoContent)
}
