package http

// health_handler.go — endpoint de salud (/health).
//
// Propósito: Docker/Kubernetes/monitores externos verifican que el servicio
// responde. Se mantiene DELIBERADAMENTE trivial (no hace ping a BD ni Redis).
//
// Rationale: un health-check que dependa de BD entrelaza vida-del-proceso
// con vida-de-la-BD. Si Postgres tose por 200ms, el orquestador reiniciaría
// el servicio en cascada. Para checks de dependencias existe otro endpoint
// opcional: /ready (readiness) que sí toca dependencias.
//
// Función:
//
//   HealthHandler(c *gin.Context)
//     → c.JSON(200, gin.H{"status": "ok", "service": "api", "time": now()})
//
// Nota: el search-service y el auth-service tendrán su propio /health
// (reusar mismo handler con "service" distinto).
//
// Futuro (Etapa 3):
//   - /ready: ping pg + redis + es con timeout 200ms; si alguno falla → 503.
//   - /metrics: Prometheus exporter (request_duration_seconds, etc.).
