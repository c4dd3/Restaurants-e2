package repomongo

// order.go — sub-DAO de órdenes para MongoDB.
//
// Struct:  type OrderRepoMongo struct { coll *mongo.Collection }
// Verify:  var _ ports.OrderRepository = (*OrderRepoMongo)(nil)
//
// Colección: orders (NO sharded en Etapa 2 — se podría shardear por
// restaurant_id o user_id si escala lo requiere).
//
// Modelado — decisión:
//   Embebemos `items` DENTRO del documento order. En Mongo esto es idiomático
//   para 1:N pequeño (una orden tiene pocos items, se leen juntos siempre).
//
//   Documento típico:
//     {
//       "_id": "uuid",
//       "user_id": "...",
//       "restaurant_id": "...",
//       "reservation_id": null | "...",
//       "items": [
//         { "product_id": "...", "quantity": 2, "unit_price": 8.5 },
//         ...
//       ],
//       "total": 17.0,
//       "pickup": true,
//       "status": "pending",
//       "created_at": ...,
//       "updated_at": ...
//     }
//
// Ventaja sobre Postgres:
//   - Sin transacción — una sola InsertOne es atómica por documento.
//   - Lectura con 1 query.
//
// Métodos:
//
// Create(ctx, o) error
//   Generar uuid.
//   r.coll.InsertOne(ctx, o)
//   ⚠ Atomicidad garantizada a nivel de documento único.
//
// FindByID(ctx, id) (*domain.Order, error)
//   FindOne({_id: id}). El documento ya trae `items` embebidos.
//
// Observación de diseño:
//   Esta asimetría (embeber en Mongo, tabla separada en pg) es deliberada —
//   cada motor se usa con su estilo idiomático, mientras el port sigue siendo
//   el mismo. Los structs de dominio tienen los tags db (para pg) y bson (para mongo)
//   bien puestos, así el mismo struct sirve para ambos.
