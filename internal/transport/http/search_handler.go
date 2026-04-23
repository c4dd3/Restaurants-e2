package http

// search_handler.go — handlers de /search/*.
//
// IMPORTANTE: Este handler LO USAN DOS BINARIOS:
//   - search-service (cmd/search/main.go) lo registra en SU router.
//   - api-service (cmd/api/main.go) NO lo registra. Si el frontend llama
//     /search/products al api-service, el api devolverá 404. La idea es que
//     el gateway/LB enrute /search/* al search-service.
//
// Alternativa simple para Etapa 2: que el api-service haga proxy/forward a
// search-service. No es elegante pero evita exponer múltiples URLs al frontend.
//
// Struct:
//
//   type SearchHandler struct {
//       idx ports.SearchIndex   // ElasticSearch adapter
//   }
//
// Rutas:
//
//   GET /search/products
//   ──────────────────────
//   1. query := c.Query("q")
//      if query == "" → 400.
//   2. limit := min(c.DefaultQuery("limit", "20"), 50)
//      category := c.Query("category")  // opcional, filtro
//   3. Si category != "" → idx.SearchByCategory(ctx, category, query, limit)
//      else               → idx.SearchProducts(ctx, query, limit)
//   4. err → http; 200 con {"items": results, "query": query}.
//
//   POST /search/reindex   (admin)
//   ──────────────────────
//   1. role := c.GetString("role"); if role != "admin" → 403.
//   2. Este endpoint dispara una reindexación completa:
//        a. El search-service hace stream de productos desde la BD primaria
//           (pg o mongo vía el service layer).
//        b. BulkIndexProducts en batches de 1000.
//        c. Devuelve {"indexed": N, "duration_ms": X}.
//   3. Ojo: es una operación bloqueante y cara. En producción se dispararía
//      como job asíncrono y se respondería 202 Accepted + job_id. Para Etapa 2
//      es síncrono (el enunciado así lo permite).
//
// Errores típicos de ES:
//   - elastic: no nodes available → 503 "search_unavailable".
//   - index not found → 500 + log (inicialización incorrecta).
//
// Rate limiting:
//   - /search/products sin auth es tentador para scraping. Considerar cap de
//     queries por IP (middleware futuro). Fuera de Etapa 2.
