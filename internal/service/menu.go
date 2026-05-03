package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// MenuService gestiona menús y sus productos asociados.
// La creación de menú + productos no es atómica en el service — la atomicidad
// es responsabilidad del adapter (Postgres usa BEGIN/COMMIT, Mongo usa transacción).
type MenuService struct {
	menus       ports.MenuRepository
	restaurants ports.RestaurantRepository
	products    ports.ProductRepository
	cache       ports.Cache
}

// NewMenuService construye el servicio inyectando sus dependencias.
func NewMenuService(
	menus ports.MenuRepository,
	restaurants ports.RestaurantRepository,
	products ports.ProductRepository,
	cache ports.Cache,
) *MenuService {
	return &MenuService{
		menus:       menus,
		restaurants: restaurants,
		products:    products,
		cache:       cache,
	}
}

// Create crea un menú con sus productos. Solo admins.
// Valida que el restaurante exista antes de persistir.
func (s *MenuService) Create(ctx context.Context, userRole string, req domain.CreateMenuRequest) (*domain.Menu, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	// Validar referencia al restaurante.
	if _, err := s.restaurants.FindByID(ctx, req.RestaurantID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrValidation
		}
		return nil, err
	}

	m := &domain.Menu{
		ID:           uuid.New().String(),
		RestaurantID: req.RestaurantID,
		Name:         req.Name,
		Description:  req.Description,
	}

	if err := s.menus.Create(ctx, m); err != nil {
		return nil, err
	}

	// Crear los productos del menú uno a uno.
	for _, preq := range req.Products {
		p := &domain.Product{
			ID:           uuid.New().String(),
			MenuID:       m.ID,
			RestaurantID: req.RestaurantID,
			Name:         preq.Name,
			Description:  preq.Description,
			Category:     preq.Category,
			Price:        preq.Price,
			Available:    preq.Available,
		}
		if err := s.products.Create(ctx, p); err != nil {
			return nil, err
		}
		m.Products = append(m.Products, *p)
	}

	_ = s.cache.DelByPattern(ctx, "products:*")
	return m, nil
}

// GetByID devuelve un menú por ID con cache-aside (TTL 5 min).
func (s *MenuService) GetByID(ctx context.Context, id string) (*domain.Menu, error) {
	var m domain.Menu
	if err := s.cache.Get(ctx, "menus:id:"+id, &m); err == nil {
		return &m, nil
	}

	result, err := s.menus.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, "menus:id:"+id, result, 5*time.Minute)
	return result, nil
}

// Update modifica name/description y, opcionalmente, reemplaza los productos.
// Solo admins. Invalida el caché del menú y de todos los productos.
func (s *MenuService) Update(ctx context.Context, userRole, id string, req domain.UpdateMenuRequest) (*domain.Menu, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	m, err := s.menus.Update(ctx, id, &req)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Del(ctx, "menus:id:"+id)
	_ = s.cache.DelByPattern(ctx, "products:*")
	return m, nil
}

// Delete elimina un menú y sus productos (ON DELETE CASCADE en Postgres).
// Solo admins.
func (s *MenuService) Delete(ctx context.Context, userRole, id string) error {
	if userRole != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	if err := s.menus.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.cache.Del(ctx, "menus:id:"+id)
	_ = s.cache.DelByPattern(ctx, "products:*")
	return nil
}
