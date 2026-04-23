package repomongo

// restaurant.go — sub-DAO de restaurantes para MongoDB.
//
// Struct:  type RestaurantRepoMongo struct { coll *mongo.Collection }
// Verify:  var _ ports.RestaurantRepository = (*RestaurantRepoMongo)(nil)
//
// Colección: restaurants  (NO sharded — poca cantidad, lecturas frecuentes por ID).
//
// Métodos:
//
// Create(ctx, r) error
//   Generar uuid si r.ID == "".
//   Fijar CreatedAt, UpdatedAt = time.Now().
//   _, err := r.coll.InsertOne(ctx, r)
//
// FindByID(ctx, id) (*domain.Restaurant, error)
//   FindOne({_id: id}); mongo.ErrNoDocuments → nil, nil.
//
// FindAll(ctx) ([]domain.Restaurant, error)
//   cur, err := r.coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"created_at": -1}))
//   defer cur.Close(ctx)
//   var out []domain.Restaurant
//   err = cur.All(ctx, &out)
//   Si out == nil → devolver []domain.Restaurant{} (evita nil checks upstream).
//
// Consideraciones:
//   - Como no está sharded, las queries son directas al shard primario.
//   - Si el catálogo crece mucho, se podría shardear por _id (hashed) — evaluar
//     con monitoreo real.
