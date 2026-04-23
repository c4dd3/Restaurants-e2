package repopg

// order.go — sub-DAO de órdenes para Postgres.
//
// Struct:  type OrderRepoPg struct { pool *pgxpool.Pool }
// Verify:  var _ ports.OrderRepository = (*OrderRepoPg)(nil)
//
// Tablas involucradas: orders + order_items.
//   orders:      id, user_id, restaurant_id, reservation_id (nullable), total,
//                pickup, status, created_at, updated_at.
//   order_items: id, order_id (FK), product_id (FK), quantity, unit_price.
//
// CRÍTICO: las operaciones que escriben order + order_items deben ser
// TRANSACCIONALES para mantener consistencia.
//
// Create(ctx, o) error — TRANSACCIÓN:
//
//   tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
//   if err != nil { return err }
//   defer tx.Rollback(ctx)   // no-op si ya se hizo Commit
//
//   // 1. Insertar orders
//   INSERT INTO orders (id, user_id, restaurant_id, reservation_id, total, pickup, status)
//   VALUES ($1, ..., $7);
//
//   // 2. Insertar order_items en batch
//   Para cada item:
//     INSERT INTO order_items (id, order_id, product_id, quantity, unit_price)
//     VALUES ($1, $2, $3, $4, $5);
//   Optimización: pgx.Batch → una sola llamada al servidor para N inserts.
//
//   // 3. Commit
//   return tx.Commit(ctx)
//
// FindByID(ctx, id) (*domain.Order, error)
//   1. SELECT * FROM orders WHERE id = $1.
//   2. SELECT * FROM order_items WHERE order_id = $1.
//   3. Asignar o.Items = items.
//   (O una sola query con LEFT JOIN + aggregation.)
//
// Observaciones:
//   - El TOTAL se persiste (no se recalcula en cada lectura). El precio
//     unitario SE CONGELA al momento de crear la orden → si el producto
//     sube de precio después, la orden histórica mantiene su importe real.
//   - reservation_id es opcional: una orden puede ser pickup sin reserva.
