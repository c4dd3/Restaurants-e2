// Command search es el microservicio de búsqueda construido sobre ElasticSearch.
//
// Endpoints previstos (se implementarán en transport/http/search):
//
//	GET  /search/products?q=texto
//	GET  /search/products/category/:categoria
//	POST /search/reindex
//
// Este servicio lee productos del repositorio principal (para reindexar)
// y los escribe en ElasticSearch. No tiene su propia BD persistente —
// ElasticSearch es un índice derivado de la fuente de verdad.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/adapters/repomongo"
	"restaurants-e2/internal/adapters/searches"
	"restaurants-e2/internal/config"
	httptransport "restaurants-e2/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[search] config inválida: %v", err)
	}

	if cfg.Engine != config.EngineMongo {
		log.Fatalf("[search] por ahora el reindex está conectado a MongoDB; use DB_ENGINE=mongo")
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	esClient, err := searches.NewClient(cfg.Search)
	if err != nil {
		log.Fatalf("[search] no se pudo conectar a ElasticSearch: %v", err)
	}

	index, err := searches.NewIndex(startupCtx, esClient, cfg.Search.Index)
	if err != nil {
		log.Fatalf("[search] no se pudo preparar el índice: %v", err)
	}

	mongoClient, err := repomongo.NewClient(startupCtx, cfg.Mongo)
	if err != nil {
		log.Fatalf("[search] no se pudo conectar a MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()

	productRepo := repomongo.NewProductRepository(
		mongoClient.Database(cfg.Mongo.DBName).Collection("products"),
	)

	log.Printf("[search] ES=%s index=%s MongoDB=%s", cfg.Search.URL, cfg.Search.Index, cfg.Mongo.DBName)

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

	addr := fmt.Sprintf(":%d", cfg.HTTP.SearchPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("[search] escuchando en %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("[search] server error: %v", err)
	}
}
