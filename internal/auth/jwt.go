package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"restaurants-e2/internal/domain"
)

// Claims son los datos del usuario que viajan dentro del JWT.
// Embebe jwt.RegisteredClaims para que el parser valide exp, iat, iss automáticamente.
type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Sign emite un JWT HS256 firmado con el secret dado.
// ttl define cuánto tiempo es válido el token (p.ej. 24h).
func Sign(userID, email, role, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "restaurants-e2",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// Parse valida y decodifica un token JWT. Devuelve domain.ErrUnauthorized si
// el token es inválido, está vencido, o usa un algoritmo inesperado.
func Parse(tokenString, secret string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		// Rechazar cualquier algoritmo que no sea HMAC (previene el ataque alg=none).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inválido")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	c, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, domain.ErrUnauthorized
	}
	return c, nil
}
