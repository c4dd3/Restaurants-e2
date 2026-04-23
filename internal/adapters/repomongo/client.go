package repomongo

// client.go — constructor del *mongo.Client.
//
// Función pública:
//
//   NewClient(ctx context.Context, cfg config.MongoConfig) (*mongo.Client, error)
//
// Lógica:
//   1. opts := options.Client().ApplyURI(cfg.URI).
//      URI típica: "mongodb://mongos:27017/restaurants"  (apunta al mongos).
//   2. Tuning:
//        opts.SetMaxPoolSize(50)
//        opts.SetMinPoolSize(5)
//        opts.SetServerSelectionTimeout(5 * time.Second)
//        opts.SetReadPreference(readpref.PrimaryPreferred())
//   3. client, err := mongo.Connect(ctx, opts).
//   4. Ping con ctx timeout 5s: client.Ping(ctx, readpref.Primary()).
//   5. Devolver client — el wiring se ocupa de client.Disconnect(ctx) al shutdown.
//
// ReadPreference:
//   - PrimaryPreferred: lecturas al primario; si está caído, van a secundario.
//   - Para reportes que toleren lag: Secondary.
//   Definir esto por tipo de query en cada sub-DAO si hace falta.
//
// Consistencia:
//   Por default Mongo usa "read concern local" y "write concern majority"
//   (w:majority). Eso significa que un write solo devuelve OK cuando la
//   mayoría del replica set confirmó. Bueno para producción — un poco más lento
//   pero seguro ante caída de un secundario.
