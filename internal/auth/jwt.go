package auth

// jwt.go — emisión y verificación de JSON Web Tokens.
//
// Algoritmo: HS256 (symmetric HMAC-SHA256). Un único secret compartido entre
// los 3 servicios (auth, api, search) vía la variable de entorno JWT_SECRET.
//
// Claims personalizadas (struct a definir):
//
//   type Claims struct {
//       UserID string `json:"sub"`
//       Email  string `json:"email"`
//       Role   string `json:"role"`
//       jwt.RegisteredClaims   // embebido: iss, exp, iat, etc.
//   }
//
// Funciones públicas:
//
//   Sign(claims Claims, secret string, ttl time.Duration) (string, error)
//     1. claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(ttl))
//     2. claims.IssuedAt  = jwt.NewNumericDate(time.Now())
//     3. claims.Issuer    = "restaurants-e2"
//     4. tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//     5. return tok.SignedString([]byte(secret))
//
//   Parse(tokenString, secret string) (*Claims, error)
//     1. parser := jwt.NewParser()
//     2. Definir la función keyFunc que valida el método y devuelve []byte(secret):
//          keyFunc := func(t *jwt.Token) (any, error) {
//            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
//              return nil, errors.New("método inválido")
//            }
//            return []byte(secret), nil
//          }
//     3. parser.ParseWithClaims(tokenString, &Claims{}, keyFunc).
//     4. Si !tok.Valid → error "token inválido".
//     5. Si c, ok := tok.Claims.(*Claims); !ok → error.
//     6. Devolver *c.
//
//   Validaciones automáticas del parser: expiración (exp), not-before (nbf),
//   issued-at (iat). Si el token está vencido, Parse devuelve error.
//
// Uso típico:
//   - auth-service: Sign al finalizar login/register.
//   - api-service:  Parse en AuthMiddleware → poner *Claims en gin.Context.
//   - search-service: idem api-service (middleware para /search/reindex).
//
// Rotación de secret (fuera de Etapa 2 pero conviene tenerlo en la cabeza):
//   Para rotar el secret sin invalidar todos los tokens vivos, se soportan
//   dos secrets simultáneos ("current" + "previous"). El parse prueba con
//   ambos. Los tokens nuevos solo se firman con "current".
