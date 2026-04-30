package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// ReservationService gestiona reservas de mesa.
//
// INVARIANTE CRÍTICA: no deben existir dos reservas confirmadas que superen la
// capacidad del restaurante en la misma ventana de tiempo.
// Esta invariante la garantiza la BD (EXCLUDE CONSTRAINT en Postgres,
// transacción optimista en Mongo). El service solo interpreta el resultado.
type ReservationService struct {
	reservations ports.ReservationRepository
	restaurants  ports.RestaurantRepository
	cache        ports.Cache
}

// NewReservationService construye el servicio inyectando sus dependencias.
func NewReservationService(
	reservations ports.ReservationRepository,
	restaurants ports.RestaurantRepository,
	cache ports.Cache,
) *ReservationService {
	return &ReservationService{
		reservations: reservations,
		restaurants:  restaurants,
		cache:        cache,
	}
}

// Create intenta crear una reserva para el usuario dado.
// Retorna domain.ErrValidation si el restaurante no existe.
// Retorna domain.ErrConflict si no hay asientos disponibles.
func (s *ReservationService) Create(ctx context.Context, userID string, req domain.CreateReservationRequest) (*domain.Reservation, error) {
	// 1. Validar que el restaurante exista.
	if _, err := s.restaurants.FindByID(ctx, req.RestaurantID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrValidation
		}
		return nil, err
	}

	// 2. Chequear disponibilidad.
	available, err := s.reservations.CheckAvailability(ctx, req.RestaurantID, req.PartySize)
	if err != nil {
		return nil, err
	}
	if available < req.PartySize {
		return nil, domain.ErrConflict
	}

	// 3. Construir y persistir la reserva.
	r := &domain.Reservation{
		ID:           uuid.New().String(),
		RestaurantID: req.RestaurantID,
		UserID:       userID,
		Date:         req.Date,
		PartySize:    req.PartySize,
		Status:       domain.StatusPending,
		Notes:        req.Notes,
	}

	if err := s.reservations.Create(ctx, r); err != nil {
		// Si el adapter detectó un conflicto de concurrencia (EXCLUDE CONSTRAINT),
		// lo devuelve como domain.ErrConflict — lo propagamos tal cual.
		return nil, err
	}

	// 4. Invalidar caché de disponibilidad de este restaurante.
	_ = s.cache.DelByPattern(ctx, "reservations:rest:"+req.RestaurantID+":*")

	return r, nil
}

// Cancel cancela la reserva si pertenece al usuario dado.
// Retorna domain.ErrForbidden si la reserva es de otro usuario.
func (s *ReservationService) Cancel(ctx context.Context, userID, reservationID string) error {
	r, err := s.reservations.FindByID(ctx, reservationID)
	if err != nil {
		return err
	}

	if r.UserID != userID {
		return domain.ErrForbidden
	}

	if err := s.reservations.Cancel(ctx, reservationID); err != nil {
		return err
	}

	_ = s.cache.DelByPattern(ctx, "reservations:rest:"+r.RestaurantID+":*")
	return nil
}
