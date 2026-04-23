package http

// user_handler.go — handlers de /users/me.
//
// Struct:
//
//   type UserHandler struct {
//       svc *service.UserService
//   }
//
// Rutas (todas protegidas por AuthMiddleware):
//
//   GET /users/me
//   ──────────────────────
//   1. uid := c.GetString("user_id")   ← lo puso AuthMiddleware
//   2. u, err := h.svc.GetMe(c.Request.Context(), uid)
//   3. if err != nil { s, b := domainToHTTP(err); c.JSON(s, b); return }
//   4. c.JSON(200, u)  ← User sin PasswordHash (domain.User.MarshalJSON o DTO)
//
//   PATCH /users/me
//   ──────────────────────
//   1. uid := c.GetString("user_id")
//   2. var req domain.UpdateUserRequest; bind.
//   3. u, err := h.svc.UpdateMe(c.Request.Context(), uid, req)
//   4. map err → http; si ok → c.JSON(200, u).
//
// Importante:
//   - El campo user_id SIEMPRE viene del token, NUNCA del body. Si un cliente
//     intenta PATCH /users/me con {"id": "otro-uuid"} debe ser ignorado.
//   - Si más adelante se agrega GET /users/:id, validar que c.Param("id") ==
//     c.GetString("user_id") o que role == "admin".
//
// Por qué PATCH y no PUT:
//   - PATCH = actualización parcial (el cliente puede enviar solo name o solo
//     password). PUT exigiría el objeto completo, que es incómodo para updates
//     parciales. REST puro discute esto, pero PATCH es lo pragmático en 2026.
