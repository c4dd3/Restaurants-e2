package repomongo

// user.go — sub-DAO de usuarios para MongoDB.
//
// Struct:  type UserRepoMongo struct { coll *mongo.Collection }
// Verify:  var _ ports.UserRepository = (*UserRepoMongo)(nil)
//
// Colección: users
// Índice recomendado (crearlo en init-cluster.sh o en boot del adapter):
//   db.users.createIndex({ email: 1 }, { unique: true })
//
// Estrategia de ID:
//   Usamos el mismo string UUID que Postgres en el campo _id. Esto mantiene
//   la simetría: mismo ID en ambos motores. Los structs de dominio tienen
//   tag `bson:"_id,omitempty"` en el campo ID.
//
// Métodos:
//
// FindByID(ctx, id) (*domain.User, error)
//   filter := bson.M{"_id": id}
//   err := r.coll.FindOne(ctx, filter).Decode(&u)
//   if errors.Is(err, mongo.ErrNoDocuments) → return nil, nil
//
// FindByEmail(ctx, email) (*domain.User, error)
//   filter := bson.M{"email": email}
//   FindOne + Decode.
//
// Create(ctx, u) error
//   Si u.ID == "" → generar uuid.
//   _, err := r.coll.InsertOne(ctx, u)
//   Si mongo.WriteException con code 11000 (duplicate key) → error envuelto;
//   el service lo traduce a ErrConflict.
//
// Update(ctx, id, req) (*domain.User, error)
//   update := bson.M{"$set": bson.M{}}
//   Solo incluir campos no vacíos:
//     if req.Name != ""  → update["$set"].(bson.M)["name"] = req.Name
//     if req.Email != "" → update["$set"].(bson.M)["email"] = req.Email
//     update["$set"].(bson.M)["updated_at"] = time.Now()
//   r.coll.FindOneAndUpdate(ctx, filter, update,
//       options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&u)
//
// Delete(ctx, id) error
//   res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
//   if res.DeletedCount == 0 → devolver nil o ErrNotFound según criterio.
//
// Observación:
//   La colección users NO se shardea (poca cardinalidad, muchas lecturas por email).
//   Solo products y reservations están sharded.
