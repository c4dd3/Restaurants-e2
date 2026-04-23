package http

// order_handler.go — handlers de /orders.
//
// Struct:
//
//   type OrderHandler struct {
//       svc *service.OrderService
//   }
//
// Rutas:
//
//   POST /orders
//   ──────────────────────
//   1. var req domain.CreateOrderRequest; bind.
//      (restaurant_id, items[] con {product_id, qty, unit_price})
//      NOTA: El unit_price del cliente es SUGERENCIA. El service lo ignora
//      y consulta el precio real del producto en BD (anti-tampering).
//   2. uid := c.GetString("user_id")
//   3. o, err := h.svc.Create(ctx, uid, req)
//   4. err → http:
//        - ErrNotFound si algún product_id no existe.
//        - ErrValidation si qty <= 0 o items vacío.
//   5. 201 con order (incluyendo total calculado server-side).
//
//   GET /orders
//   ──────────────────────
//   1. uid := c.GetString("user_id")
//   2. list, err := h.svc.ListByUser(ctx, uid)
//   3. err → http; 200.
//
//   GET /orders/:id
//   ──────────────────────
//   1. Ownership check idéntico a reservations.
//   2. err → http; 200.
//
//   PATCH /orders/:id/status   (admin / staff del restaurante)
//   ──────────────────────
//   1. var req struct { Status string `json:"status" binding:"required,oneof=preparing ready delivered cancelled"` }
//      bind.
//   2. Validar role == admin (staff-multi-role es scope futuro).
//   3. o, err := h.svc.UpdateStatus(ctx, c.Param("id"), req.Status, role)
//   4. err → http; 200.
//
// Notas:
//   - El total SIEMPRE se calcula en el service usando precios reales de la BD.
//     Nunca confiar en unit_price ni total del cliente.
//   - El service arma la orden en TX (order + order_items en pg; embebido en mongo).
//   - Si el producto dejó de existir entre el GET menú y el POST order,
//     devolver ErrNotFound → 404 con detail "product_not_found".
