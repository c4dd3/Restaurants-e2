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

	"restaurants-e2/internal/adapters/repopg"
	"restaurants-e2/internal/config"
	"restaurants-e2/internal/ports"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[api] config inválida: %v", err)
	}
	log.Printf("[api] motor de BD seleccionado: %s", cfg.Engine)

	ctx := context.Background()

	repos, cleanup, err := buildRepositories(ctx, cfg)
	if err != nil {
		log.Fatalf("[api] no se pudo conectar a la base de datos: %v", err)
	}
	defer cleanup()

	// TODO: instanciar services con repos y registrar rutas en transport/http.
	_ = repos

	gin.SetMode(cfg.HTTP.GinMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "api",
			"status":  "ok",
			"engine":  cfg.Engine,
		})
	})

	addr := fmt.Sprintf(":%d", cfg.HTTP.APIPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	runWithGracefulShutdown(srv, "api")
}

// buildRepositories conecta al motor de BD elegido y devuelve los repos listos.
// cleanup debe llamarse con defer para liberar conexiones al apagar el servidor.
func buildRepositories(ctx context.Context, cfg *config.Config) (*ports.Repositories, func(), error) {
	switch cfg.Engine {
	case config.EnginePostgres:
		pool, err := repopg.NewPool(ctx, cfg.Postgres)
		if err != nil {
			return nil, nil, err
		}
		repos := repopg.NewRepositories(pool)
		return repos, pool.Close, nil

	case config.EngineMongo:
		// TODO: instanciar repomongo cuando esté implementado.
		return nil, nil, errors.New("motor mongo aún no implementado")

	default:
		return nil, nil, fmt.Errorf("motor desconocido: %s", cfg.Engine)
	}
}

// runWithGracefulShutdown levanta el servidor y atiende señales de sistema
// para hacer shutdown ordenado — evita cortar requests en vuelo.
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
