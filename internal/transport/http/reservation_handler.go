package http

// reservation_handler.go — handlers de /reservations.
//
// Struct:
//
//   type ReservationHandler struct {
//       svc *service.ReservationService
//   }
//
// Rutas:
//
//   POST /reservations
//   ──────────────────────
//   1. var req domain.CreateReservationRequest; bind.
//      (restaurant_id, date, people_count)
//   2. uid := c.GetString("user_id")
//   3. r, err := h.svc.Create(ctx, uid, req)
//   4. err → http:
//        - ErrConflict (409) → "slot_not_available" (colisión de horario).
//        - ErrValidation (422) → fecha pasada, people_count inválido, etc.
//   5. 201 con reservation.
//
//   GET /reservations  (listar las propias)
//   ──────────────────────
//   1. uid := c.GetString("user_id")
//   2. list, err := h.svc.ListByUser(ctx, uid)
//   3. err → http; 200.
//
//   GET /reservations/:id
//   ──────────────────────
//   1. id := c.Param("id")
//   2. uid := c.GetString("user_id"); role := c.GetString("role")
//   3. r, err := h.svc.GetByID(ctx, id, uid, role)
//      ← el service verifica ownership (o admin). Si no corresponde → 403.
//   4. err → http; 200.
//
//   DELETE /reservations/:id  (cancelar)
//   ──────────────────────
//   1. Similar ownership check.
//   2. Service actualiza status a "cancelled" (no delete físico, para auditoría).
//   3. 204.
//
// Notas sobre disponibilidad:
//   - El service hace CheckAvailability ANTES de INSERT. Aún así, dos
//     requests simultáneos podrían ambos pasar el check. La mitigación
//     depende del motor:
//       - Postgres: constraint EXCLUDE USING gist (descrito en repopg/reservation.go).
//       - Mongo: unique index parcial + retry tras duplicate key error.
//   - El handler NO se preocupa de concurrencia; solo refleja el resultado.
//
// Por qué GET /reservations solo devuelve las del usuario:
//   - Privacidad por defecto. Un admin que quiera todas usa un endpoint
//     separado (/admin/reservations) o query param explicit.
