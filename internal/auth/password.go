package auth

// password.go — wrapping de bcrypt para hashear y verificar contraseñas.
//
// Librería: golang.org/x/crypto/bcrypt.
//
// Funciones públicas:
//
//   HashPassword(plain string) (string, error)
//     1. bytes, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
//     2. return string(bytes), err
//
//   VerifyPassword(hash, plain string) error
//     1. return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
//     Devuelve nil si coinciden, error si no.
//
//   const BcryptCost = 12
//     Costo (log₂ del número de rondas). 12 ≈ 250ms por hash en una CPU moderna
//     (2026). Alto lo suficiente para resistir bruteforce; bajo lo suficiente
//     para no bloquear el servidor en cada login. Si se detectan CPUs más
//     rápidas en producción, subir a 13-14 y rehashar al próximo login.
//
// Política:
//   - NUNCA guardar la contraseña en claro (ni en logs, ni en BD, ni en caché).
//   - Política mínima: 6 caracteres (forzado por binding en RegisterRequest).
//     Idealmente 10+ para producción, pero el curso exige lo mínimo.
//   - No implementar rate limiting acá — es trabajo del API Gateway o middleware.
//     En Etapa 2 se puede agregar un middleware de rate limiting simple basado
//     en IP con go-redis (opcional).
//
// Errores relevantes:
//   - bcrypt.ErrMismatchedHashAndPassword  → el service lo convierte a ErrInvalidCredentials.
//   - bcrypt.ErrHashTooShort / ErrCost     → indica hash corrupto en BD (bug).
