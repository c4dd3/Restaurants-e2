package cacheredis

// client.go — constructor del *redis.Client.
//
// Función pública:
//
//   NewClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error)
//
// Lógica:
//   opts := &redis.Options{
//       Addr:     cfg.Addr,        // ej: "redis:6379"
//       Password: cfg.Password,    // vacío en desarrollo
//       DB:       cfg.DB,          // típicamente 0
//       PoolSize: 10,
//       MinIdleConns: 2,
//   }
//   client := redis.NewClient(opts)
//   if err := client.Ping(ctx).Err(); err != nil { return nil, err }
//   return client, nil
//
// El caller (wiring) se ocupa de client.Close() en el shutdown.
