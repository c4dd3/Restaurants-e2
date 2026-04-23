package wiring

// repositories.go — el ÚNICO switch/if del proyecto que decide el motor de BD.
//
// Función pública:
//
//   NewRepositories(ctx context.Context, cfg *config.Config) (*ports.Repositories, CloseFunc, error)
//
// Donde CloseFunc es:
//
//   type CloseFunc func(ctx context.Context) error
//
// Lógica:
//
//   switch cfg.Engine {
//   case config.EnginePostgres:
//       pool, err := repopg.NewPool(ctx, cfg.Postgres)
//       if err != nil { return nil, nil, err }
//       repos := repopg.NewRepositories(pool)
//       closer := func(_ context.Context) error { pool.Close(); return nil }
//       return repos, closer, nil
//
//   case config.EngineMongo:
//       client, err := repomongo.NewClient(ctx, cfg.Mongo)
//       if err != nil { return nil, nil, err }
//       repos := repomongo.NewRepositories(client, cfg.Mongo.Database)
//       closer := func(ctx context.Context) error { return client.Disconnect(ctx) }
//       return repos, closer, nil
//
//   default:
//       return nil, nil, fmt.Errorf("engine no soportado: %q", cfg.Engine)
//   }
//
// Usos:
//   cmd/api/main.go   → repos, close, err := wiring.NewRepositories(ctx, cfg)
//   cmd/auth/main.go  → idem
//   cmd/search/main.go → NO llama esto (el search-service no habla con la BD
//                        primaria; solo con ES).
//
// Este archivo materializa la intención del enunciado: "un solo lugar decide
// la BD; todo el resto solo ve interfaces". Si algún día se agrega un tercer
// motor (p.ej. CockroachDB), acá se agrega un case y el resto del código ni
// se inmuta.
