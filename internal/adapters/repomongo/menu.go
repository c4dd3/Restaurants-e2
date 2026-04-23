package repomongo

// menu.go — sub-DAO de menús para MongoDB.
//
// Struct:  type MenuRepoMongo struct { coll *mongo.Collection }
// Verify:  var _ ports.MenuRepository = (*MenuRepoMongo)(nil)
//
// Colección: menus (NO sharded).
//
// Decisión de modelado:
//   Opción A — menú con productos REFERENCIADOS (otra colección):
//     menus:    { _id, restaurant_id, name, description }
//     products: { _id, menu_id, name, category, price, ... }
//     ✓ Consistente con Postgres.
//     ✓ Se puede buscar productos directamente por categoría (la shard key).
//     ✗ Requiere 2 queries para leer menú con productos.
//
//   Opción B — menú con productos EMBEBIDOS:
//     menus: { _id, restaurant_id, name, products: [ {name, price, ...}, ... ] }
//     ✓ Lectura en 1 query.
//     ✗ No se puede shardear products por categoría.
//     ✗ Búsqueda por producto se complica.
//
// Elegimos Opción A para mantener simetría con Postgres y habilitar el sharding
// de products.
//
// Métodos:
//
// Create(ctx, m) error
//   _, err := r.coll.InsertOne(ctx, m)
//
// FindByID(ctx, id) (*domain.Menu, error)
//   1. FindOne({_id: id}).Decode(&menu)
//   2. cur, _ := productsColl.Find(ctx, bson.M{"menu_id": id})
//      cur.All(ctx, &menu.Products)
//   Si el repo solo tiene menusColl, inyectar también productsColl por constructor
//   o resolverlo desde r.coll.Database().Collection("products").
//
// Update(ctx, id, req) (*domain.Menu, error)
//   FindOneAndUpdate con $set.
//   Si req.Products != nil → transacción:
//     session, _ := client.StartSession()
//     session.WithTransaction(ctx, func(sessCtx) {
//       DeleteMany products con menu_id = id.
//       InsertMany productos nuevos.
//     })
//
// Delete(ctx, id) error
//   TX: borrar menú + borrar products asociados.
//   En Mongo NO hay ON DELETE CASCADE — se simula en código.
