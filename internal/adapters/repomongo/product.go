package repomongo

// product.go — sub-DAO de productos para MongoDB.
//
// Struct:  type ProductRepoMongo struct { coll *mongo.Collection }
// Verify:  var _ ports.ProductRepository = (*ProductRepoMongo)(nil)
//
// Colección: products
// Sharding: SÍ — shard key "category" (hashed).
//   Configurado en deployments/mongo/init-cluster.sh:
//     sh.shardCollection("restaurants.products", { category: "hashed" })
//
// Implicancias del sharding:
//   - Escrituras distribuidas: productos de categorías distintas caen en
//     shards distintos → paralelismo de inserción.
//   - FindByCategory: el mongos sabe a qué shard ir (targeted query).
//   - FindByID: el mongos NO sabe qué shard tiene el id → broadcast a todos.
//     Mitigación: incluir category en el filter si se conoce; si no, aceptar
//     el broadcast (aceptable en Etapa 2 — no es hot path).
//   - Cada INSERT DEBE incluir el campo category (shard key obligatoria).
//
// Métodos:
//
// FindByID(ctx, id) (*domain.Product, error)
//   FindOne({_id: id}). Si no existe → nil, nil.
//
// FindByCategory(ctx, category) ([]domain.Product, error)
//   filter := bson.M{"category": category}
//   Find + cur.All.
//   Targeted: va al shard que contiene ese hash de category.
//
// FindAll(ctx) ([]domain.Product, error)
//   filter := bson.M{}
//   ⚠ Broadcast a todos los shards — evitar sin paginación.
//
// Create(ctx, p) error
//   Validar p.Category != "" (shard key obligatoria).
//   InsertOne.
//
// Update(ctx, p) error
//   ⚠ En Mongo sharded, NO se puede cambiar el valor de la shard key.
//   Si el service necesita cambiar la categoría, debe DeleteOne + InsertOne
//   (o devolver error). Para Etapa 2: disallow en el service y devolver ErrValidation.
//
// Delete(ctx, id) error
//   Si se conoce la category → incluirla: DeleteOne({_id: id, category: cat}).
//   Si no → DeleteOne({_id: id}) con broadcast. Aceptable.
