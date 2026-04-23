package http

// menu_handler.go — handlers de /menus.
//
// Struct:
//
//   type MenuHandler struct {
//       svc *service.MenuService
//   }
//
// Rutas:
//
//   POST /menus   (admin)
//   ──────────────────────
//   1. var req domain.CreateMenuRequest; bind.
//      (Incluye restaurant_id, name, products[] con ProductRequest embebido.)
//   2. role := c.GetString("role")
//   3. m, err := h.svc.Create(ctx, req, role)
//   4. err → http; 201 con menu (incluyendo IDs de productos creados).
//
//   GET /menus/:id
//   ──────────────────────
//   1. id := c.Param("id")
//   2. m, err := h.svc.GetByID(ctx, id)
//      ← el service puede incluir los productos (JOIN en pg, $lookup en mongo).
//   3. err → http; 200.
//
//   GET /restaurants/:id/menus
//   ──────────────────────
//   (Endpoint opcional, útil para listar todos los menús de un restaurante.)
//   1. restaurantID := c.Param("id")
//   2. list, err := h.svc.ListByRestaurant(ctx, restaurantID)
//   3. err → http; 200.
//
//   PATCH /menus/:id   (admin)
//   ──────────────────────
//   UpdateMenuRequest. Validar role. Service ejecuta.
//
//   DELETE /menus/:id  (admin)
//   ──────────────────────
//   Cascadea a productos si el modelado lo exige. 204 si ok.
//
// Notas:
//   - El bind en POST /menus hace validación profunda (cada ProductRequest
//     tiene sus tags binding:"required,min,max"). Si un producto del array
//     es inválido → 400 antes de tocar el service.
//   - El service ejecuta esto en TX: el menú y sus productos son atómicos.
//     Si cualquier producto falla, no se crea nada.
