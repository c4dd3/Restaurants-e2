package repomongo

// reservation.go — sub-DAO de reservas para MongoDB.
//
// Struct:  type ReservationRepoMongo struct { coll *mongo.Collection }
// Verify:  var _ ports.ReservationRepository = (*ReservationRepoMongo)(nil)
//
// Colección: reservations
// Sharding: SÍ — shard key "restaurant_id" (hashed).
//   Configurado en deployments/mongo/init-cluster.sh:
//     sh.shardCollection("restaurants.reservations", { restaurant_id: "hashed" })
//
// Por qué esa shard key:
//   Las consultas "availability de un restaurante X" son frecuentes. Al
//   shardear por restaurant_id hashed, todas las reservas de un mismo
//   restaurante viven en el MISMO shard → CheckAvailability golpea 1 shard
//   (no broadcast). El hash balancea la distribución entre shards.
//
// Índice complementario (crear en init):
//   db.reservations.createIndex({ restaurant_id: 1, date: 1, status: 1 })
//   → acelera las queries de disponibilidad por ventana de tiempo.
//
// Métodos:
//
// Create(ctx, r) error
//   Validar r.RestaurantID != "" (shard key obligatoria).
//   InsertOne.
//   ⚠ No hay exclusion constraint como en Postgres — la validación
//   de solapamiento se hace a nivel service (CheckAvailability antes del insert).
//   Para mayor garantía bajo concurrencia: usar TX + findOneAndUpdate con
//   filtro de capacidad disponible (optimistic concurrency).
//
// FindByID(ctx, id) (*domain.Reservation, error)
//   ⚠ Broadcast (no sabemos el restaurant_id). Mitigación: incluirlo en
//   el filter si el service lo conoce.
//
// Cancel(ctx, id) error
//   UpdateOne({_id: id, status: {$ne: "cancelled"}},
//             {$set: {status: "cancelled", updated_at: time.Now()}})
//   Chequear ModifiedCount == 0 → ErrNotFound o ya estaba cancelada.
//
// CheckAvailability(ctx, restaurantID, partySize) (int, error)
//   Lógica:
//     1. restaurantsColl.FindOne({_id: restaurantID}).Decode(&rest)  // capacity
//     2. pipeline := [
//          {$match: {
//            restaurant_id: restaurantID,
//            status: "confirmed",
//            date: {$gte: now, $lt: now+2h}
//          }},
//          {$group: {_id: null, total: {$sum: "$party_size"}}}
//        ]
//        cur, _ := coll.Aggregate(ctx, pipeline)
//        // extraer "total" — si cur.TryNext → total=0
//     3. available := rest.Capacity - total
//     4. return available
//
//   Todo va a UN solo shard gracias a la shard key. Rápido.
