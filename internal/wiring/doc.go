// Package wiring concentra la construcción de dependencias que cruzan varias
// capas. Es el ÚNICO lugar del proyecto donde se decide qué motor de BD usar.
//
// Por qué existe este paquete:
//
//	El requisito del curso dice "solo UNA condición (if) elige el motor".
//	Hay tres binarios (cmd/api, cmd/auth, cmd/search) que necesitan repos.
//	Si cada main tuviera su propio `switch cfg.Engine`, habría tres switches.
//	Centralizar en `wiring.NewRepositories` cumple literal el requisito y
//	previene que una divergencia en el wiring de dos binarios rompa algo sutil.
//
// Reglas:
//
//  1. wiring/ puede importar adapters/ directamente (es la excepción al
//     "nadie fuera de main importa adapters"). wiring ES un main auxiliar.
//  2. wiring/ NO importa transport/http ni service — es agnóstico de la
//     capa superior; solo arma lo que está por debajo.
//  3. Los constructores devuelven io.Closer si hace falta para que el main
//     pueda hacer defer close() al shutdown.
package wiring
