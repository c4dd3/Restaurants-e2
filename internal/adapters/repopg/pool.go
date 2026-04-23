package repopg

// pool.go — constructor del pool de conexiones Postgres (pgx/v5).
//
// Función pública a implementar:
//
//   NewPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error)
//
// Lógica:
//   1. Armar DSN usando cfg.ResolvedDSN()  (ya existe en internal/config).
//   2. pgxpool.ParseConfig(dsn) → *pgxpool.Config.
//   3. Tunear el pool:
//        poolCfg.MaxConns        = 10   (o 5×CPU cores del contenedor)
//        poolCfg.MinConns        = 2
//        poolCfg.MaxConnIdleTime = 5 * time.Minute
//        poolCfg.MaxConnLifetime = 30 * time.Minute
//        poolCfg.HealthCheckPeriod = 1 * time.Minute
//      Estos números son conservadores; ajustar con pg_stat_activity si hace falta.
//   4. pgxpool.NewWithConfig(ctx, poolCfg).
//   5. Ping con ctx timeout 5s. Si falla → propagar error envuelto.
//   6. Devolver el pool — el caller (wiring) se ocupa de cerrarlo con defer pool.Close().
//
// Por qué pgx/v5 y no database/sql + lib/pq:
//   - pgx es el driver nativo más performante (~30% más rápido en benchmarks).
//   - API tipada con pgxpool.Pool (no interface{} por todos lados).
//   - Protocolo extended por default (prepared statements gratis).
//   - Scan directo a structs con RowToStructByName — sin ORM.
