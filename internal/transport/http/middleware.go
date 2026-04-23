package http

// middleware.go — middlewares HTTP transversales.
//
// Lista:
//
//   1. RequestID()                    — genera un UUID por request y lo
//                                        inyecta en el contexto y en el header
//                                        X-Request-ID. Ayuda a correlacionar
//                                        logs con errores reportados por
//                                        cliente.
//
//   2. Logger()                       — loguea método, path, status, latencia,
//                                        user_id (si está), request_id.
//                                        Formato JSON (one line per request),
//                                        amigable para ingestarse en ES/Loki.
//
//   3. CORS()                         — habilita CORS para desarrollo. En
//                                        producción ajustar AllowedOrigins a
//                                        dominios del frontend. Usar
//                                        github.com/gin-contrib/cors.
//
//   4. AuthMiddleware(secret string)  — valida el JWT del header
//                                        Authorization: Bearer <token>.
//                                        Pasos:
//                                          a. leer header; si falta o no tiene
//                                             prefijo "Bearer " → 401.
//                                          b. claims, err := auth.Parse(tok, secret)
//                                             → si err → 401.
//                                          c. c.Set("user_id", claims.UserID)
//                                             c.Set("role",    claims.Role)
//                                             c.Set("email",   claims.Email)
//                                          d. c.Next()
//
//   5. AdminOnly()                    — middleware opcional sobre AuthMiddleware;
//                                        si c.GetString("role") != "admin" → 403.
//                                        Útil en rutas como POST /restaurants
//                                        o /search/reindex.
//
// Firma de un middleware en Gin:
//
//   func NombreDelMiddleware() gin.HandlerFunc {
//       return func(c *gin.Context) {
//           // pre
//           c.Next()
//           // post (opcional)
//       }
//   }
//
// Orden importa en el router:
//
//   r.Use(Recovery, RequestID, Logger, CORS)  // siempre
//   grupoProtegido.Use(AuthMiddleware(secret)) // solo en rutas autenticadas
//
// Por qué NO usar sessions/cookies:
//   El enunciado exige JWT. Además, tres servicios stateless comparten el
//   secret; con cookies se requeriría un store compartido (Redis).
