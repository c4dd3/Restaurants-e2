package http

// auth_handler.go — handlers de /auth/register y /auth/login.
//
// Struct:
//
//   type AuthHandler struct {
//       svc *service.AuthService
//   }
//
//   func NewAuthHandler(svc *service.AuthService) *AuthHandler
//
// Rutas:
//
//   POST /auth/register
//   ──────────────────────
//   1. var req domain.RegisterRequest
//      if err := c.ShouldBindJSON(&req); err != nil {
//          c.JSON(400, gin.H{"error": "bad_request", "detail": err.Error()})
//          return
//      }
//      (El binding valida: email format, min=6 en password, name required.)
//   2. resp, err := h.svc.Register(c.Request.Context(), req)
//   3. if err != nil { s, b := domainToHTTP(err); c.JSON(s, b); return }
//   4. c.JSON(201, resp)  ← 201 Created (devuelve token + user)
//
//   POST /auth/login
//   ──────────────────────
//   1. var req domain.LoginRequest
//      if err := c.ShouldBindJSON(&req); err != nil {
//          c.JSON(400, gin.H{"error": "bad_request"})
//          return
//      }
//   2. resp, err := h.svc.Login(c.Request.Context(), req)
//   3. if err != nil { s, b := domainToHTTP(err); c.JSON(s, b); return }
//   4. c.JSON(200, resp)  ← 200 OK (devuelve token + user)
//
// Notas de seguridad:
//   - El service NO distingue "email inexistente" de "contraseña incorrecta"
//     → ambos devuelven ErrInvalidCredentials → 401. Evita user enumeration.
//   - Nunca loguear req.Password (ni si falla el bind).
//   - Content-Type esperado: application/json. Si el cliente envía form-urlencoded,
//     ShouldBindJSON falla con 400.
//
// Por qué 201 en register y 200 en login:
//   - Register CREA un recurso nuevo (usuario) → 201 es semánticamente correcto.
//   - Login solo autentica → 200.
