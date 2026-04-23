package searches

// client.go — constructor del cliente de ElasticSearch.
//
// Función pública:
//
//   NewClient(cfg config.SearchConfig) (*elasticsearch.Client, error)
//
// Lógica:
//   cfg := elasticsearch.Config{
//       Addresses: []string{cfg.URL},           // ej: "http://elasticsearch:9200"
//       // En desarrollo: xpack.security.enabled=false → sin auth.
//       // En prod: agregar Username/Password o APIKey.
//   }
//   es, err := elasticsearch.NewClient(cfg)
//   if err != nil { return nil, err }
//
//   // Smoke test: GET /
//   res, err := es.Info()
//   res.Body.Close()
//   if res.IsError() → return nil, fmt.Errorf("ES info: %s", res.Status())
//   return es, nil
//
// Boot-up recomendado (hacerlo una vez en el main del search-service):
//   - Si el índice "products" no existe, crearlo con el mapping correcto
//     (ver doc.go).
//   - La función EnsureIndex(ctx, client) *Index se puede implementar acá
//     o en index.go — elegir una.
