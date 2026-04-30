package auth

import "golang.org/x/crypto/bcrypt"

// BcryptCost define el costo de hashing.
// 12 ≈ 250 ms por hash en una CPU moderna (2026). Suficiente para resistir
// bruteforce sin bloquear el servidor en cada login.
const BcryptCost = 12

// HashPassword genera un hash bcrypt de la contraseña en texto claro.
// El resultado es seguro para almacenar en base de datos.
func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword compara un hash con una contraseña en texto claro.
// Devuelve nil si coinciden, error si no.
// El caller debe convertir bcrypt.ErrMismatchedHashAndPassword → domain.ErrInvalidCredentials.
func VerifyPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
