package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// RestaurantService gestiona restaurantes con cache-aside en Redis.
type RestaurantService struct {
	restaurants ports.RestaurantRepository
	cache       ports.Cache
}

// NewRestaurantService construye el servicio inyectando sus dependencias.
func NewRestaurantService(restaurants ports.RestaurantRepository, cache ports.Cache) *RestaurantService {
	return &RestaurantService{restaurants: restaurants, cache: cache}
}

// Create crea un restaurante. Solo admins pueden hacerlo.
// userID se usa para establecer AdminID del restaurante.
func (s *RestaurantService) Create(ctx context.Context, userID, userRole string, req domain.CreateRestaurantRequest) (*domain.Restaurant, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	r := &domain.Restaurant{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Address:     req.Address,
		Phone:       req.Phone,
		Description: req.Description,
		AdminID:     userID,
		Capacity:    req.Capacity,
	}

	if err := s.restaurants.Create(ctx, r); err != nil {
		return nil, err
	}

	// Invalidar listados cacheados — ya no están actualizados.
	_ = s.cache.DelByPattern(ctx, "restaurants:*")

	return r, nil
}

// GetByID devuelve un restaurante por ID con cache-aside (TTL 5 min).
func (s *RestaurantService) GetByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	var r domain.Restaurant
	if err := s.cache.Get(ctx, "restaurants:id:"+id, &r); err == nil {
		return &r, nil
	}

	result, err := s.restaurants.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, "restaurants:id:"+id, result, 5*time.Minute)
	return result, nil
}

// List devuelve todos los restaurantes con cache-aside (TTL 5 min).
func (s *RestaurantService) List(ctx context.Context) ([]domain.Restaurant, error) {
	var list []domain.Restaurant
	if err := s.cache.Get(ctx, "restaurants:all", &list); err == nil {
		return list, nil
	}

	result, err := s.restaurants.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, "restaurants:all", result, 5*time.Minute)
	return result, nil
}
