package config

import (
	"os"
	"testing"
)

// TestLoadDefaults verifica que Load funciona con la configuración mínima (JWT_SECRET).
func TestLoadDefaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "super-secret-key-123")
	t.Setenv("DB_ENGINE", "postgres")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load falló con config válida: %v", err)
	}
	if cfg.Engine != EnginePostgres {
		t.Fatalf("Engine esperado 'postgres', obtuvo %q", cfg.Engine)
	}
	if cfg.JWT.Secret != "super-secret-key-123" {
		t.Fatalf("JWT.Secret no cargado correctamente")
	}
	// Valores por defecto
	if cfg.Postgres.Host != "localhost" {
		t.Fatalf("Postgres.Host esperado 'localhost', obtuvo %q", cfg.Postgres.Host)
	}
	if cfg.Redis.Addr != "localhost:6379" {
		t.Fatalf("Redis.Addr esperado 'localhost:6379', obtuvo %q", cfg.Redis.Addr)
	}
}

func TestLoadMongoEngine(t *testing.T) {
	t.Setenv("JWT_SECRET", "super-secret-key-123")
	t.Setenv("DB_ENGINE", "mongo")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load falló con engine=mongo: %v", err)
	}
	if cfg.Engine != EngineMongo {
		t.Fatalf("Engine esperado 'mongo', obtuvo %q", cfg.Engine)
	}
}

func TestLoadInvalidEngine(t *testing.T) {
	t.Setenv("JWT_SECRET", "super-secret-key-123")
	t.Setenv("DB_ENGINE", "sqlite")

	_, err := Load()
	if err == nil {
		t.Fatal("Load debió fallar con DB_ENGINE inválido")
	}
}

func TestValidateMissingSecret(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate debió fallar con JWT_SECRET vacío")
	}
}

func TestValidateSecretTooShort(t *testing.T) {
	cfg := &Config{JWT: JWTConfig{Secret: "corto"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate debió fallar con JWT_SECRET < 16 chars")
	}
}

func TestValidateOK(t *testing.T) {
	cfg := &Config{JWT: JWTConfig{Secret: "sixteen-chars-ok!"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate falló con secret válido: %v", err)
	}
}

func TestResolvedDSNExplicit(t *testing.T) {
	p := PostgresConfig{DSN: "postgres://user:pass@host:5432/db"}
	if got := p.ResolvedDSN(); got != "postgres://user:pass@host:5432/db" {
		t.Fatalf("ResolvedDSN con DSN explícita incorrecto: %q", got)
	}
}

func TestResolvedDSNBuilt(t *testing.T) {
	p := PostgresConfig{Host: "db", Port: 5432, User: "u", Password: "p", DBName: "mydb"}
	got := p.ResolvedDSN()
	expected := "postgres://u:p@db:5432/mydb?sslmode=disable"
	if got != expected {
		t.Fatalf("ResolvedDSN construida incorrecta: got %q want %q", got, expected)
	}
}

func TestGetEnvFallback(t *testing.T) {
	os.Unsetenv("TEST_KEY_XYZ")
	if v := getEnv("TEST_KEY_XYZ", "default"); v != "default" {
		t.Fatalf("getEnv sin var debió retornar fallback, obtuvo %q", v)
	}
}

func TestGetEnvSet(t *testing.T) {
	t.Setenv("TEST_KEY_XYZ", "valor")
	if v := getEnv("TEST_KEY_XYZ", "default"); v != "valor" {
		t.Fatalf("getEnv con var seteada debió retornar %q, obtuvo %q", "valor", v)
	}
}

func TestGetEnvAsIntFallback(t *testing.T) {
	os.Unsetenv("TEST_INT_XYZ")
	if v := getEnvAsInt("TEST_INT_XYZ", 42); v != 42 {
		t.Fatalf("getEnvAsInt sin var debió retornar 42, obtuvo %d", v)
	}
}

func TestGetEnvAsIntSet(t *testing.T) {
	t.Setenv("TEST_INT_XYZ", "99")
	if v := getEnvAsInt("TEST_INT_XYZ", 0); v != 99 {
		t.Fatalf("getEnvAsInt con var '99' debió retornar 99, obtuvo %d", v)
	}
}

func TestGetEnvAsIntInvalid(t *testing.T) {
	t.Setenv("TEST_INT_XYZ", "noesunentero")
	if v := getEnvAsInt("TEST_INT_XYZ", 7); v != 7 {
		t.Fatalf("getEnvAsInt con valor no-int debió retornar fallback 7, obtuvo %d", v)
	}
}
