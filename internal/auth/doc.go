// Package auth concentra los helpers de SEGURIDAD compartidos entre
// auth-service (emite credenciales) y api-service (las verifica).
//
// Archivos:
//
//	jwt.go       ← firmar y parsear JSON Web Tokens (HS256).
//	password.go  ← hashear y verificar passwords con bcrypt.
//
// Por qué este paquete existe:
//   Tener DOS servicios que emiten/verifican tokens exige que ambos usen
//   EXACTAMENTE el mismo algoritmo, formato de claims y secret. Si cada
//   servicio tuviera su propia lógica, una divergencia mínima (p.ej. un
//   claim renombrado) rompería la autenticación sin dar un error claro.
//   Este paquete es "la única fuente de verdad" de cómo se emiten/verifican
//   credenciales en todo el sistema.
//
// Seguridad — principios:
//   1. Nada en este paquete escribe logs con el secret, la contraseña, o el token.
//   2. Todas las funciones son puras (no I/O, no BD). Son trivialmente testeables.
//   3. Los errores son genéricos para el exterior (no filtrar "password incorrecta"
//      vs "usuario inexistente" — eso lo decide el service a nivel de UX).
package auth
