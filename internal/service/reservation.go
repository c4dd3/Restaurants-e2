package service

// ReservationService — casos de uso sobre reservas.
//
// INVARIANTE CRÍTICA:
//   No deben existir dos reservas confirmadas que se solapen para el mismo
//   restaurante y que superen su capacidad.
//
// Dependencias:
//   - ports.ReservationRepository
//   - ports.RestaurantRepository    (para chequear capacidad)
//   - ports.Cache
//
// Métodos públicos:
//
//   Create(ctx, userID string, req CreateReservationRequest) (*domain.Reservation, error)
//     1. Validar que el restaurante exista (RestaurantRepository.FindByID).
//        Si no → ErrValidation.
//     2. Chequear disponibilidad:
//        - ReservationRepository.CheckAvailability(ctx, restaurant_id, party_size).
//        - Devuelve cantidad de asientos disponibles considerando reservas
//          activas en la ventana de tiempo solicitada.
//        - Si availableSeats < party_size → ErrConflict.
//     3. Construir domain.Reservation (uuid, user_id, restaurant_id, status=pending).
//     4. ReservationRepository.Create(ctx, &r).
//     5. Cache.DelByPattern(ctx, "reservations:rest:"+restaurant_id+":*").
//     6. Devolver la reserva.
//
//   Cancel(ctx, userID, reservationID string) error
//     1. ReservationRepository.FindByID.
//        - Si no existe → ErrNotFound.
//        - Si la reserva no pertenece al userID → ErrForbidden.
//     2. ReservationRepository.Cancel(ctx, id).  (cambia status a cancelled)
//     3. Invalidar caché.
//
// Shard key en Mongo:
//   La colección `reservations` se shardea por restaurant_id (hashed).
//   Eso permite que las queries CheckAvailability golpeen UN solo shard
//   (las reservas de un mismo restaurante viven juntas).
//
// Race condition (importante defender en la oral):
//   Entre el paso 2 (CheckAvailability) y el paso 4 (Create), dos requests
//   concurrentes podrían "ganar" los mismos asientos.
//   Mitigación por BD:
//     - Postgres → EXCLUSION CONSTRAINT con tstzrange && + btree_gist,
//       garantiza a nivel de BD que no hay solapamientos.
//     - Mongo    → transacción ACID sobre el replica set + findOneAndUpdate
//       condicional (optimistic concurrency).
//   El service confía en que el adapter cumple la invariante: si recibe un
//   error de unicidad/conflicto del adapter, lo mapea a ErrConflict.
