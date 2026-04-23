package service

// AuthService — casos de uso de autenticación.
//
// Es el ÚNICO lugar del sistema que maneja contraseñas en claro.
// Una vez hasheadas y persistidas, nadie más las ve.
//
// Dependencias (inyectadas por constructor):
//   - ports.UserRepository  (sub-DAO elegido por wiring según DB_ENGINE)
//   - auth.Hasher           (bcrypt wrapper — internal/auth/password.go)
//   - auth.TokenSigner      (JWT wrapper — internal/auth/jwt.go)
//
// Métodos públicos:
//
//   Register(ctx, RegisterRequest) (*domain.User, string, error)
//     1. Validar el DTO (Gin ya valida "required, email, min=6" por tags, pero
//        acá chequeamos reglas de negocio: email único, role ∈ {client, admin}).
//     2. Verificar que no exista un usuario con ese email.
//         - UserRepository.FindByEmail → si devuelve != nil → ErrConflict.
//     3. Hashear la contraseña: auth.HashPassword(pwd) → string.
//     4. Construir domain.User con ID (uuid) + email + hash + role + timestamps.
//     5. UserRepository.Create(ctx, &u).
//     6. Firmar JWT con (user_id, email, role, exp=24h).
//     7. Devolver usuario + token.
//
//   Login(ctx, LoginRequest) (*domain.User, string, error)
//     1. UserRepository.FindByEmail(ctx, email) → si no existe → ErrInvalidCredentials.
//        ⚠ Devolver SIEMPRE el mismo error (no distinguir "email inexistente" de
//        "password errónea") — evita enumeración de cuentas.
//     2. auth.VerifyPassword(hash, pwd) → si falla → ErrInvalidCredentials.
//     3. Firmar JWT y devolver.
//
// Notas de seguridad:
//   - Bcrypt con cost ≥ 12 (ajustable por config; default seguro para 2026).
//   - JWT con HS256; secret compartido solo con api-service y search-service.
//   - Expiración del token: 24h por default (configurable).
//   - No loguear la contraseña ni el hash bajo ningún concepto.
