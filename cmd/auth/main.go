// Command auth es el microservicio de autenticación.
// Emite tokens JWT firmados con HS256 usando JWT_SECRET.
// El api-service valida tokens usando el mismo secreto, sin llamadas HTTP cruzadas
// (stateless auth).
//
// Endpoints previstos (se implementarán en transport/http/auth):
//   POST /auth/register
//   POST /auth/login
//
// Este servicio solo necesita acceso al UserRepository (no a menús/órdenes).
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
		log.Fatalf("[auth] config inválida: %v", err)
	}

	log.Printf("[auth] motor de BD seleccionado: %s", cfg.Engine)

	gin.SetMode(cfg.HTTP.GinMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "auth", "status": "ok"})
	})

	// TODO:
	//   - Inyectar UserRepository (postgres|mongo según cfg.Engine).
	//   - Registrar POST /auth/register y POST /auth/login.

	addr := fmt.Sprintf(":%d", cfg.HTTP.AuthPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("[auth] escuchando en %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("[auth] server error: %v", err)
	}
}
