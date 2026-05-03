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

	"restaurants-e2/internal/adapters/cacheredis"
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
		log.Fatalf("[api] config inválida: %v", err)
	}
	log.Printf("[api] motor de BD seleccionado: %s", cfg.Engine)

	ctx := context.Background()

	// ── Repositorios ─────────────────────────────────────────────────────────
	repos, cleanupRepos, err := buildRepositories(ctx, cfg)
	if err != nil {
		log.Fatalf("[api] no se pudo conectar a la base de datos: %v", err)
	}
	defer cleanupRepos()

	// ── Caché Redis ───────────────────────────────────────────────────────────
	redisClient, err := cacheredis.NewClient(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("[api] no se pudo conectar a Redis: %v", err)
	}
	defer redisClient.Close()
	cache := cacheredis.New(redisClient)

	// ── Services ──────────────────────────────────────────────────────────────
	userSvc := service.NewUserService(repos.Users)
	restaurantSvc := service.NewRestaurantService(repos.Restaurants, cache)
	menuSvc := service.NewMenuService(repos.Menus, repos.Restaurants, repos.Products, cache)
	productSvc := service.NewProductService(repos.Products, cache)
	reservationSvc := service.NewReservationService(repos.Reservations, repos.Restaurants, cache)
	orderSvc := service.NewOrderService(repos.Orders, repos.Products, repos.Restaurants)

	// ── Router ────────────────────────────────────────────────────────────────
	gin.SetMode(cfg.HTTP.GinMode)
	r := transport.NewRouter(transport.Deps{
		UserService:        userSvc,
		RestaurantService:  restaurantSvc,
		MenuService:        menuSvc,
		ProductService:     productSvc,
		ReservationService: reservationSvc,
		OrderService:       orderSvc,
		JWTSecret:          cfg.JWT.Secret,
	})

	// ── Servidor HTTP con graceful shutdown ───────────────────────────────────
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
		return repopg.NewRepositories(pool), pool.Close, nil

	case config.EngineMongo:
		client, err := repomongo.NewClient(ctx, cfg.Mongo)
		if err != nil {
			return nil, nil, err
		}

		repos := repomongo.NewRepositories(client, cfg.Mongo.DBName)

		cleanup := func() {
			_ = client.Disconnect(context.Background())
		}

		return repos, cleanup, nil

	default:
		return nil, nil, fmt.Errorf("motor desconocido: %s", cfg.Engine)
	}
}

// runWithGracefulShutdown levanta el servidor y atiende señales SIGINT/SIGTERM
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
