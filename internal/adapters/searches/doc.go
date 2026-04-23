// Package searches implementa el port ports.SearchIndex contra ElasticSearch
// (go-elasticsearch/v8).
//
// Tipo exportado: *Index. Al pie de index.go:
//
//	var _ ports.SearchIndex = (*Index)(nil)
//
// Responsabilidades:
//   - Convertir domain.Product ↔ documentos ES.
//   - Traducir queries del service en consultas DSL de ES (match, term, bool).
//   - Manejar el bulk API para reindexaciones masivas.
//
// Estrategia de índice:
//   - Un solo índice: "products".
//   - _id del documento = id del producto (idempotencia en reindex).
//   - Mapping explícito con:
//       name:        text + keyword multifield (búsqueda + exact match)
//       description: text
//       category:    keyword
//       price:       double
//       available:   boolean
//     Se crea en el primer boot (CreateIndex si no existe).
package searches
