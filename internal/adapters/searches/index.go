package searches

// index.go — implementación de ports.SearchIndex.
//
// Struct:
//   type Index struct {
//       es        *elasticsearch.Client
//       indexName string   // típicamente "products"
//   }
// Verify:
//   var _ ports.SearchIndex = (*Index)(nil)
//
// Métodos:
//
// IndexProduct(ctx, p) error
//   Si p.Description == "" → setear "Producto sin descripción" ANTES de serializar
//   (en el índice queremos el texto searchable, no el vacío).
//   body, _ := json.Marshal(p)
//   req := esapi.IndexRequest{
//       Index:      idx.indexName,
//       DocumentID: p.ID,
//       Body:       bytes.NewReader(body),
//       Refresh:    "false",   // eventual consistency; no bloquear al writer
//   }
//   res, err := req.Do(ctx, idx.es)
//   defer res.Body.Close()
//   if res.IsError() → return fmt.Errorf("index: %s", res.String())
//
// BulkIndexProducts(ctx, ps) error
//   Armar el payload NDJSON del bulk API:
//     { "index": { "_index": "products", "_id": "<id>" } }\n
//     { ...product json... }\n
//     ... por cada producto ...
//   Chunks de 500 productos por request para no saturar.
//   Evaluar la respuesta: si "errors": true → inspeccionar items[] y
//   devolver lista de fallos (logs + error agregado).
//
// SearchProducts(ctx, query, limit) ([]domain.Product, error)
//   Query DSL:
//     {
//       "size": limit,
//       "query": {
//         "multi_match": {
//           "query": query,
//           "fields": ["name^3", "description", "category^2"],
//           "fuzziness": "AUTO"
//         }
//       }
//     }
//   ^3 y ^2 son boosts: name pesa más que category, que pesa más que description.
//   fuzziness AUTO tolera typos de 1-2 letras.
//   Decode response → _source por hit → slice de domain.Product.
//
// SearchByCategory(ctx, category, limit) ([]domain.Product, error)
//   Query DSL:
//     {
//       "size": limit,
//       "query": { "term": { "category": category } }
//     }
//   term (exact match) sobre el keyword — O(1) en el inverted index.
//
// DeleteProduct(ctx, id) error
//   esapi.DeleteRequest{Index: idx.indexName, DocumentID: id}
//   Si 404 → no-op (idempotente).
//
// Consideraciones:
//   - refresh="false": ES refresca el índice cada 1s por default → los docs
//     nuevos son visibles en ~1s. Bueno para throughput.
//   - Para consistencia fuerte (test de "inserté y busco enseguida") usar
//     refresh="wait_for" — costo en latencia.
//   - El search-service NO depende de Postgres/Mongo directamente — todo lo
//     que indexa lo recibe vía HTTP del api-service (POST /search/reindex).
