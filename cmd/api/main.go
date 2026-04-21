// Command api es el microservicio principal: CRUD de restaurantes, menús,
// productos, reservas y órdenes. Es agnóstico del motor de BD — se elige
// vía DB_ENGINE.
//
// Responsabilidades:
//   - Cargar config, elegir adaptador Postgres o Mongo.
//   - Inicializar cliente Redis (caché) y router Gin.
//   - Exponer /health para que el balanceador sepa cuándo incluir esta réplica.
//   - Shutdown elegante al recibir SIGTERM (importante en contenedores).
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

	"restaurants-e2/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[api] config inválida: %v", err)
	}

	// TODO(wiring):
	//   - Según cfg.Engine, construir un ports.Repositories con repopg o repomongo.
	//   - Instanciar cacheredis.New(cfg.Redis) y pasarlo al layer de servicio.
	//   - Crear services con repos + cache, y registrar rutas en transport/http.
	//
	// Este scaffolding levanta el servidor con solo /health para validar que la infra funciona.
	log.Printf("[api] motor de BD seleccionado: %s", cfg.Engine)

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

// runWithGracefulShutdown levanta el servidor y atiende señales de sistema
// para hacer shutdown ordenado — evita cortar requests en vuelo cuando el
// contenedor se reinicia (importante con escalado horizontal).
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
