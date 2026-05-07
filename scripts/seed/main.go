package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"restaurants-e2/internal/adapters/repopg"
	"restaurants-e2/internal/auth"
	"restaurants-e2/internal/config"
	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

func main() {
	// ── Flags ──────────────────────────────────────────────────────────────
	numRestaurants := flag.Int("restaurants", 10, "cantidad de restaurantes a crear")
	menusPerRest := flag.Int("menus-per", 2, "menús por restaurante")
	productsPerMenu := flag.Int("products-per", 8, "productos por menú")
	numUsers := flag.Int("users", 20, "usuarios normales a crear")
	adminEmail := flag.String("admin-email", "admin@seed.com", "email del admin")
	adminPassword := flag.String("admin-password", "admin12345", "password del admin")
	dryRun := flag.Bool("dry-run", false, "genera datos pero NO inserta en BD")
	flag.Parse()

	// ── Config ─────────────────────────────────────────────────────────────
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[seed] config inválida: %v", err)
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	var llm LLMClient
	if apiKey == "" {
		log.Printf("[seed] ANTHROPIC_API_KEY no configurada — usando datos estáticos realistas")
		llm = &StaticClient{}
	} else {
		model := os.Getenv("SEED_LLM_MODEL")
		if model == "" {
			model = "claude-haiku-4-5-20251001"
		}
		llm = NewAnthropicClient(apiKey, model)
	}
	ctx := context.Background()

	// ── Repositorios ────────────────────────────────────────────────────────
	var repos *ports.Repositories
	if !*dryRun {
		repos, err = connectRepos(ctx, cfg)
		if err != nil {
			log.Fatalf("[seed] no se pudo conectar a la BD: %v", err)
		}
		log.Printf("[seed] conectado a %s", cfg.Engine)
	}

	stats := struct{ users, restaurants, menus, products int }{}

	// ── Admin ───────────────────────────────────────────────────────────────
	log.Printf("[seed] creando admin %s...", *adminEmail)
	hashed, err := auth.HashPassword(*adminPassword)
	if err != nil {
		log.Fatalf("[seed] error hasheando password: %v", err)
	}
	admin := &domain.User{
		ID:        uuid.NewString(),
		Name:      "Admin Seed",
		Email:     *adminEmail,
		Password:  hashed,
		Role:      domain.RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if !*dryRun {
		if err := repos.Users.Create(ctx, admin); err != nil {
			log.Printf("[seed] admin ya existe o error: %v", err)
		}
	}

	// ── Usuarios ────────────────────────────────────────────────────────────
	log.Printf("[seed] generando %d usuarios con LLM...", *numUsers)
	users, err := generateUsers(ctx, llm, *numUsers)
	if err != nil {
		log.Fatalf("[seed] error generando usuarios: %v", err)
	}
	for _, u := range users {
		hashed, _ := auth.HashPassword("password123")
		u.Password = hashed
		u.ID = uuid.NewString()
		u.Role = domain.RoleClient
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
		if *dryRun {
			log.Printf("  [dry] usuario: %s <%s>", u.Name, u.Email)
		} else {
			if err := repos.Users.Create(ctx, &u); err != nil {
				log.Printf("  [warn] usuario %s: %v", u.Email, err)
				continue
			}
		}
		stats.users++
	}
	log.Printf("[seed] %d usuarios creados", stats.users)

	// ── Restaurantes + Menús + Productos ────────────────────────────────────
	log.Printf("[seed] generando %d restaurantes con LLM...", *numRestaurants)
	restaurants, err := generateRestaurants(ctx, llm, *numRestaurants)
	if err != nil {
		log.Fatalf("[seed] error generando restaurantes: %v", err)
	}

	for i, r := range restaurants {
		r.ID = uuid.NewString()
		r.AdminID = admin.ID
		r.CreatedAt = time.Now()
		r.UpdatedAt = time.Now()

		if *dryRun {
			log.Printf("  [dry] restaurante: %s (%s)", r.Name, r.Address)
		} else {
			if err := repos.Restaurants.Create(ctx, &r); err != nil {
				log.Printf("  [warn] restaurante %s: %v", r.Name, err)
				continue
			}
		}
		stats.restaurants++

		// Menús para este restaurante
		for m := 0; m < *menusPerRest; m++ {
			log.Printf("[seed] restaurante %d/%d — menú %d/%d...", i+1, len(restaurants), m+1, *menusPerRest)
			menu, products, err := generateMenuWithProducts(ctx, llm, r.Name, r.Description, *productsPerMenu)
			if err != nil {
				log.Printf("  [warn] menú para %s: %v", r.Name, err)
				continue
			}

			menu.ID = uuid.NewString()
			menu.RestaurantID = r.ID
			menu.CreatedAt = time.Now()
			menu.UpdatedAt = time.Now()

			if *dryRun {
				log.Printf("  [dry] menú: %s (%d productos)", menu.Name, len(products))
			} else {
				if err := repos.Menus.Create(ctx, menu); err != nil {
					log.Printf("  [warn] menú %s: %v", menu.Name, err)
					continue
				}
			}
			stats.menus++

			for _, p := range products {
				p.ID = uuid.NewString()
				p.MenuID = menu.ID
				p.RestaurantID = r.ID
				if p.Description == "" {
					p.Description = domain.DefaultProductDescription
				}
				if *dryRun {
					log.Printf("    [dry] producto: %s (₡%.0f)", p.Name, p.Price)
				} else {
					if err := repos.Products.Create(ctx, &p); err != nil {
						log.Printf("    [warn] producto %s: %v", p.Name, err)
						continue
					}
				}
				stats.products++
			}
		}

		// Pequeña pausa para no quemar la quota del LLM
		time.Sleep(500 * time.Millisecond)
	}

	// ── Resumen ─────────────────────────────────────────────────────────────
	fmt.Printf("\n✅ Seed completado%s:\n", func() string {
		if *dryRun {
			return " (dry-run)"
		}
		return ""
	}())
	fmt.Printf("   Usuarios:     %d\n", stats.users)
	fmt.Printf("   Restaurantes: %d\n", stats.restaurants)
	fmt.Printf("   Menús:        %d\n", stats.menus)
	fmt.Printf("   Productos:    %d\n", stats.products)
}

// ── Helpers de generación ────────────────────────────────────────────────────

func generateUsers(ctx context.Context, llm LLMClient, n int) ([]domain.User, error) {
	system, user := promptUsers(n, int(time.Now().Unix()))
	raw, err := llm.Complete(ctx, system, user)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Users []struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"users"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("parse usuarios: %w — raw: %s", err, raw[:min(200, len(raw))])
	}

	users := make([]domain.User, 0, len(resp.Users))
	for _, u := range resp.Users {
		users = append(users, domain.User{Name: u.Name, Email: u.Email})
	}
	return users, nil
}

func generateRestaurants(ctx context.Context, llm LLMClient, n int) ([]domain.Restaurant, error) {
	system, user := promptRestaurants(n, int(time.Now().Unix()))
	raw, err := llm.Complete(ctx, system, user)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Restaurants []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Address     string `json:"address"`
			Phone       string `json:"phone"`
			Category    string `json:"category"`
			Capacity    int    `json:"capacity"`
		} `json:"restaurants"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("parse restaurantes: %w — raw: %s", err, raw[:min(200, len(raw))])
	}

	restaurants := make([]domain.Restaurant, 0, len(resp.Restaurants))
	for _, r := range resp.Restaurants {
		restaurants = append(restaurants, domain.Restaurant{
			Name:        r.Name,
			Description: r.Description,
			Address:     r.Address,
			Phone:       r.Phone,
			Capacity:    r.Capacity,
		})
	}
	return restaurants, nil
}

func generateMenuWithProducts(ctx context.Context, llm LLMClient, restaurantName, category string, numProducts int) (*domain.Menu, []domain.Product, error) {
	system, user := promptMenuWithProducts(restaurantName, category, numProducts)
	raw, err := llm.Complete(ctx, system, user)
	if err != nil {
		return nil, nil, err
	}

	var resp struct {
		Menu struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Products    []struct {
				Name        string  `json:"name"`
				Description string  `json:"description"`
				Category    string  `json:"category"`
				Price       float64 `json:"price"`
				Available   bool    `json:"available"`
			} `json:"products"`
		} `json:"menu"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, nil, fmt.Errorf("parse menú: %w — raw: %s", err, raw[:min(200, len(raw))])
	}

	menu := &domain.Menu{
		Name:        resp.Menu.Name,
		Description: resp.Menu.Description,
	}

	products := make([]domain.Product, 0, len(resp.Menu.Products))
	for _, p := range resp.Menu.Products {
		products = append(products, domain.Product{
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Price:       p.Price,
			Available:   p.Available,
		})
	}
	return menu, products, nil
}

func connectRepos(ctx context.Context, cfg *config.Config) (*ports.Repositories, error) {
	switch cfg.Engine {
	case config.EnginePostgres:
		pool, err := repopg.NewPool(ctx, cfg.Postgres)
		if err != nil {
			return nil, err
		}
		return repopg.NewRepositories(pool), nil
	default:
		return nil, fmt.Errorf("engine %q no soportado en seed aún", cfg.Engine)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
