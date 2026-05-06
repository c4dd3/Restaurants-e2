// Command search es el microservicio de búsqueda construido sobre ElasticSearch.
//
// Endpoints:
//
//	GET  /search/products?q=texto
//	GET  /search/products/category/:categoria
//	POST /search/reindex
//
// Este servicio lee productos del repositorio principal (para reindexar)
// y los escribe en ElasticSearch. Soporta DB_ENGINE=postgres y DB_ENGINE=mongo.
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
	"restaurants-e2/internal/adapters/searches"
	"restaurants-e2/internal/config"
	"restaurants-e2/internal/ports"
	httptransport "restaurants-e2/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[search] config inválida: %v", err)
	}
	log.Printf("[search] motor de BD: %s", cfg.Engine)

	startupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ── ElasticSearch ─────────────────────────────────────────────────────
	esClient, err := searches.NewClient(cfg.Search)
	if err != nil {
		log.Fatalf("[search] no se pudo conectar a ElasticSearch: %v", err)
	}

	index, err := searches.NewIndex(startupCtx, esClient, cfg.Search.Index)
	if err != nil {
		log.Fatalf("[search] no se pudo preparar el índice: %v", err)
	}

	// ── Repositorio de productos (para reindex) ────────────────────────────
	productRepo, cleanup, err := buildProductRepo(startupCtx, cfg)
	if err != nil {
		log.Fatalf("[search] no se pudo conectar al repo de productos: %v", err)
	}
	defer cleanup()

	log.Printf("[search] ES=%s index=%s", cfg.Search.URL, cfg.Search.Index)

	// ── Router ─────────────────────────────────────────────────────────────
	gin.SetMode(cfg.HTTP.GinMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "search", "status": "ok"})
	})
	r.GET("/search/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "search", "status": "ok"})
	})

	searchHandler := httptransport.NewSearchHandler(index, productRepo)
	searchHandler.RegisterRoutes(r)

	// ── Servidor con graceful shutdown ─────────────────────────────────────
	addr := fmt.Sprintf(":%d", cfg.HTTP.SearchPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("[search] escuchando en %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("[search] señal recibida, iniciando shutdown...")
	case err := <-errCh:
		log.Fatalf("[search] server error: %v", err)
	}

	shutdownCtx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[search] shutdown falló: %v", err)
	}
	log.Println("[search] detenido limpiamente")
}

// buildProductRepo conecta al motor de BD y devuelve solo el ProductRepository.
func buildProductRepo(ctx context.Context, cfg *config.Config) (ports.ProductRepository, func(), error) {
	switch cfg.Engine {
	case config.EnginePostgres:
		pool, err := repopg.NewPool(ctx, cfg.Postgres)
		if err != nil {
			return nil, nil, err
		}
		repo := repopg.NewProductRepoPg(pool)
		return repo, pool.Close, nil

	case config.EngineMongo:
		client, err := repomongo.NewClient(ctx, cfg.Mongo)
		if err != nil {
			return nil, nil, err
		}
		cleanup := func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = client.Disconnect(shutdownCtx)
		}
		repo := repomongo.NewProductRepository(
			client.Database(cfg.Mongo.DBName).Collection("products"),
		)
		return repo, cleanup, nil

	default:
		return nil, nil, fmt.Errorf("motor desconocido: %s", cfg.Engine)
	}
}
