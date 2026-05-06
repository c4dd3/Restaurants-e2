// Command auth es el microservicio de autenticación.
// Emite tokens JWT firmados con HS256 usando JWT_SECRET.
// El api-service valida los tokens usando el mismo secreto, sin llamadas HTTP
// entre servicios (stateless auth).
//
// Nginx rutea /auth/* hacia este servicio conservando el prefijo, por lo que
// los endpoints se registran como /auth/register y /auth/login.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/adapters/repomongo"
	"restaurants-e2/internal/adapters/repopg"
	"restaurants-e2/internal/config"
	"restaurants-e2/internal/ports"
	"restaurants-e2/internal/service"
	transport "restaurants-e2/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[auth] config inválida: %v", err)
	}
	log.Printf("[auth] motor de BD seleccionado: %s", cfg.Engine)

	ctx := context.Background()

	// ── Repositorio de usuarios (único que necesita el auth-service) ──────────
	userRepo, cleanup, err := buildUserRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("[auth] no se pudo conectar a la base de datos: %v", err)
	}
	defer cleanup()

	// ── AuthService ───────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.TTL)

	// ── Router mínimo ─────────────────────────────────────────────────────────
	gin.SetMode(cfg.HTTP.GinMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), transport.RequestID())

	// Nginx no hace rewrite en /auth/* — el path llega completo al servicio.
	// Se registran ambos para que el healthcheck del Dockerfile (:8081/health)
	// y el del balanceador (/auth/health) funcionen.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "auth", "status": "ok"})
	})
	r.GET("/auth/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "auth", "status": "ok"})
	})

	ah := transport.NewAuthHandler(authSvc)
	auth := r.Group("/auth")
	{
		auth.POST("/register", ah.Register)
		auth.POST("/login", ah.Login)
	}

	// ── Servidor HTTP con graceful shutdown ───────────────────────────────────
	addr := fmt.Sprintf(":%d", cfg.HTTP.AuthPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	runWithGracefulShutdown(srv, "auth")
}

// buildUserRepository conecta al motor de BD y devuelve solo el UserRepository.
// El auth-service no necesita acceso a menús, órdenes ni reservas.
func buildUserRepository(ctx context.Context, cfg *config.Config) (ports.UserRepository, func(), error) {
	switch cfg.Engine {
	case config.EnginePostgres:
		pool, err := repopg.NewPool(ctx, cfg.Postgres)
		if err != nil {
			return nil, nil, err
		}
		return repopg.NewUserRepoPg(pool), pool.Close, nil

	case config.EngineMongo:
		client, err := repomongo.NewClient(ctx, cfg.Mongo)
		if err != nil {
			return nil, nil, err
		}

		repos := repomongo.NewRepositories(client, cfg.Mongo.DBName)

		cleanup := func() {
			_ = client.Disconnect(context.Background())
		}

		return repos.Users, cleanup, nil

	default:
		return nil, nil, fmt.Errorf("motor desconocido: %s", cfg.Engine)
	}
}

// runWithGracefulShutdown levanta el servidor y atiende señales SIGINT/SIGTERM.
func runWithGracefulShutdown(srv *http.Server, name string) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("[%s] escuchando en %s", name, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Printf("[%s] señal recibida, iniciando shutdown...", name)
	case err := <-errCh:
		log.Fatalf("[%s] server error: %v", name, err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[%s] shutdown falló: %v", name, err)
	}
	log.Printf("[%s] detenido limpiamente", name)
}
