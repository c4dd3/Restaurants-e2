package http

// product_handler.go — handlers de /products.
//
// Struct:
//
//   type ProductHandler struct {
//       svc *service.ProductService
//   }
//
// Rutas:
//
//   GET /products/:id
//   ──────────────────────
//   1. id := c.Param("id")
//   2. p, err := h.svc.GetByID(ctx, id)    ← service intenta cache primero.
//   3. err → http; 200.
//
//   GET /menus/:id/products
//   ──────────────────────
//   (Productos de un menú específico — útil para rendering en frontend.)
//   1. menuID := c.Param("id")
//   2. list, err := h.svc.ListByMenu(ctx, menuID)
//   3. err → http; 200.
//
//   POST /products   (admin)
//   ──────────────────────
//   1. var req struct {
//          MenuID string `json:"menu_id" binding:"required,uuid"`
//          Product domain.ProductRequest `json:"product" binding:"required"`
//      }
//      bind.
//   2. p, err := h.svc.Create(ctx, req.MenuID, req.Product, role)
//   3. err → http; 201.
//   4. El service indexa el producto en ES (side-effect asíncrono posible).
//
//   PATCH /products/:id  (admin)
//   ──────────────────────
//   UpdateProductRequest. Service invalida cache + reindex ES.
//
//   DELETE /products/:id (admin)
//   ──────────────────────
//   Service elimina de BD + borra de ES. 204.
//
// Notas:
//   - Los endpoints de búsqueda (/search/products) viven en search_handler.go,
//     separados porque los sirve el search-service, no el api-service.
//   - GET /products/:id usa cache-aside (TTL corto 30-60s). Handler agnóstico.
