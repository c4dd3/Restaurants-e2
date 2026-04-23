package repopg

// reservation.go — sub-DAO de reservas para Postgres.
//
// Struct:  type ReservationRepoPg struct { pool *pgxpool.Pool }
// Verify:  var _ ports.ReservationRepository = (*ReservationRepoPg)(nil)
//
// Tabla: reservations
// Columnas: id, user_id (FK), restaurant_id (FK), date (TIMESTAMPTZ), party_size,
//           notes, status, created_at, updated_at.
//
// Posible constraint de unicidad (recomendado — agregar a init.sql):
//   Impedir solapamientos a nivel BD con EXCLUDE USING gist:
//
//     CREATE EXTENSION IF NOT EXISTS btree_gist;
//     ALTER TABLE reservations
//       ADD CONSTRAINT no_overlap_per_restaurant
//       EXCLUDE USING gist (
//         restaurant_id WITH =,
//         tstzrange(date, date + interval '2 hours') WITH &&
//       ) WHERE (status != 'cancelled');
//
//   Esto garantiza a nivel de BD que dos reservas confirmadas del mismo
//   restaurante no pueden solaparse en una ventana de 2h — incluso bajo
//   concurrencia. Es la defensa final contra la race condition.
//
// Métodos:
//
// Create(ctx, r) error
//   INSERT INTO reservations (id, user_id, restaurant_id, date, party_size, notes, status)
//   VALUES ($1, ..., $7);
//   Si el INSERT viola la exclusion constraint → pgconn.PgError con SQLState
//   de tipo exclusion → devolver error envuelto; el service lo mapea a ErrConflict.
//
// FindByID(ctx, id) (*domain.Reservation, error)
//   SELECT * FROM reservations WHERE id = $1;
//
// Cancel(ctx, id) error
//   UPDATE reservations SET status = 'cancelled', updated_at = now()
//   WHERE id = $1 AND status != 'cancelled';
//   Chequear RowsAffected == 0 → ErrNotFound (o ya estaba cancelada).
//
// CheckAvailability(ctx, restaurantID, partySize) (int, error)
//   Lógica:
//     1. SELECT capacity FROM restaurants WHERE id = $1;
//     2. SELECT COALESCE(SUM(party_size), 0) FROM reservations
//        WHERE restaurant_id = $1 AND status = 'confirmed'
//          AND date >= now() AND date < now() + interval '2 hours';
//     3. available = capacity - sum_confirmed
//     4. return available.
//   Si available < partySize, el service lo traduce a ErrConflict.
//
// Nota: la ventana de "2 horas" podría parametrizarse por restaurante; para
// Etapa 2 alcanza con un valor fijo en config.
