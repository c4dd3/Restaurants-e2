package http

// restaurant_handler.go — handlers de /restaurants.
//
// Struct:
//
//   type RestaurantHandler struct {
//       svc *service.RestaurantService
//   }
//
// Rutas:
//
//   POST /restaurants   (admin only — validado por service)
//   ──────────────────────
//   1. var req domain.CreateRestaurantRequest; bind.
//   2. role := c.GetString("role")
//   3. r, err := h.svc.Create(ctx, req, role)
//   4. err → http; 201 con restaurant.
//
//   GET /restaurants
//   ──────────────────────
//   1. Query params opcionales:
//        limit  = min(c.DefaultQuery("limit", "20"), 100)   ← cap duro
//        offset = c.DefaultQuery("offset", "0")
//   2. list, err := h.svc.List(ctx, limit, offset)
//   3. err → http; 200 con {"items": list, "limit": limit, "offset": offset}.
//
//   GET /restaurants/:id
//   ──────────────────────
//   1. id := c.Param("id")
//   2. r, err := h.svc.GetByID(ctx, id)
//   3. err → http; 200 con restaurant.
//
//   PATCH /restaurants/:id   (admin)
//   ──────────────────────
//   Bind UpdateRestaurantRequest (agregar a dto.go si falta),
//   servicio valida role, limpia cache.
//
//   DELETE /restaurants/:id  (admin)
//   ──────────────────────
//   Servicio ejecuta soft-delete (DELETE físico si cascade conveniente).
//   204 No Content si éxito.
//
// Validaciones importantes:
//   - :id debe ser UUID válido. Si no, el repo devolverá ErrNotFound y el
//     cliente recibirá 404 (consistente con "no existe").
//   - Cap duro de limit=100 previene DoS accidental / scraping.
//
// Cache (decisión del service):
//   - GET /restaurants se sirve desde cache (TTL 60s). El handler ni sabe.
//   - POST/PATCH/DELETE invalidan cache en el service.
