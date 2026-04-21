// Command search es el microservicio de búsqueda construido sobre ElasticSearch.
//
// Endpoints previstos (se implementarán en transport/http/search):
//   GET  /search/products?q=texto
//   GET  /search/products/category/:categoria
//   POST /search/reindex
//
// Este servicio lee productos del repositorio principal (para reindexar)
// y los escribe en ElasticSearch. No tiene su propia BD persistente —
// ElasticSearch es un índice derivado de la fuente de verdad.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[search] config inválida: %v", err)
	}

	log.Printf("[search] ES=%s index=%s", cfg.Search.URL, cfg.Search.Index)

	gin.SetMode(cfg.HTTP.GinMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "search", "status": "ok"})
	})

	// TODO:
	//   - Inicializar cliente go-elasticsearch y crear el índice si no existe.
	//   - Inyectar ProductRepository (para /search/reindex).
	//   - Registrar las rutas /search/*.

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
