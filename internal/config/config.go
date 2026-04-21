// Package config centraliza la carga de configuración desde variables de entorno.
// El objetivo es que NINGÚN otro paquete lea directamente os.Getenv — todo pasa por acá.
// Esto facilita testear (se puede inyectar un Config falso) y documentar qué vars existen.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// DBEngine enumera los motores de base de datos soportados.
type DBEngine string

const (
	EnginePostgres DBEngine = "postgres"
	EngineMongo    DBEngine = "mongo"
)

// Config agrupa toda la configuración del sistema.
// Al agrupar por área (DB, Cache, Search, HTTP) se hace evidente qué vars influyen dónde.
type Config struct {
	Engine DBEngine

	Postgres PostgresConfig
	Mongo    MongoConfig
	Redis    RedisConfig
	Search   SearchConfig
	HTTP     HTTPConfig
	JWT      JWTConfig
	Log      LogConfig
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	DSN      string // Si está presente, sobreescribe los campos anteriores
}

// ResolvedDSN devuelve la DSN efectiva: la que el usuario pasó explícita, o una armada.
func (p PostgresConfig) ResolvedDSN() string {
	if p.DSN != "" {
		return p.DSN
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.DBName)
}

type MongoConfig struct {
	URI    string
	DBName string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
}

type SearchConfig struct {
	URL   string
	Index string
}

type HTTPConfig struct {
	APIPort    int
	AuthPort   int
	SearchPort int
	GinMode    string
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type LogConfig struct {
	Level string
}

// Load lee las variables de entorno y arma un Config listo para usar.
// Si existe un archivo .env local, lo carga primero (útil en desarrollo).
// En producción (docker-compose o k8s) las vars vienen directas del entorno.
func Load() (*Config, error) {
	// godotenv.Load no falla si el .env no existe — es solo una comodidad.
	_ = godotenv.Load()

	engine := DBEngine(getEnv("DB_ENGINE", "postgres"))
	if engine != EnginePostgres && engine != EngineMongo {
		return nil, fmt.Errorf("DB_ENGINE inválido: %q (use 'postgres' o 'mongo')", engine)
	}

	cfg := &Config{
		Engine: engine,
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnv("POSTGRES_DB", "restaurants"),
			DSN:      getEnv("POSTGRES_DSN", ""),
		},
		Mongo: MongoConfig{
			URI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
			DBName: getEnv("MONGO_DB", "restaurants"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TTL:      time.Duration(getEnvAsInt("REDIS_TTL_SECONDS", 300)) * time.Second,
		},
		Search: SearchConfig{
			URL:   getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
			Index: getEnv("ELASTICSEARCH_INDEX", "products"),
		},
		HTTP: HTTPConfig{
			APIPort:    getEnvAsInt("API_PORT", 8080),
			AuthPort:   getEnvAsInt("AUTH_PORT", 8081),
			SearchPort: getEnvAsInt("SEARCH_PORT", 8082),
			GinMode:    getEnv("GIN_MODE", "release"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			TTL:    time.Duration(getEnvAsInt("JWT_TTL_HOURS", 24)) * time.Hour,
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate aplica chequeos sanos: JWT_SECRET presente, longitudes mínimas, etc.
// Falla rápido en startup antes de aceptar tráfico.
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET es obligatorio")
	}
	if len(c.JWT.Secret) < 16 {
		return errors.New("JWT_SECRET debe tener al menos 16 caracteres")
	}
	return nil
}

// --- Helpers privados ---

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
