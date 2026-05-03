package service

import (
	"context"
	"time"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// ProductService gestiona productos individuales con cache-aside.
// Los productos se crean normalmente vía MenuService, pero este service
// permite consultarlos y modificarlos de forma independiente.
type ProductService struct {
	products ports.ProductRepository
	cache    ports.Cache
}

// NewProductService construye el servicio inyectando sus dependencias.
func NewProductService(products ports.ProductRepository, cache ports.Cache) *ProductService {
	return &ProductService{products: products, cache: cache}
}

// GetByID devuelve un producto por ID con cache-aside (TTL 10 min).
func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	if err := s.cache.Get(ctx, "products:id:"+id, &p); err == nil {
		return &p, nil
	}

	result, err := s.products.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, "products:id:"+id, result, 10*time.Minute)
	return result, nil
}

// ListByCategory devuelve los productos de una categoría con cache-aside (TTL 10 min).
// En MongoDB esta query es targeted (hit directo al shard de esa categoría).
func (s *ProductService) ListByCategory(ctx context.Context, category string) ([]domain.Product, error) {
	var list []domain.Product
	if err := s.cache.Get(ctx, "products:cat:"+category, &list); err == nil {
		return list, nil
	}

	result, err := s.products.FindByCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, "products:cat:"+category, result, 10*time.Minute)
	return result, nil
}

// Update modifica un producto. Solo admins.
// Invalida tanto la entrada individual como todos los listados por categoría
// (no sabemos si cambió de categoría, así que es más seguro limpiar todo).
func (s *ProductService) Update(ctx context.Context, userRole string, p *domain.Product) (*domain.Product, error) {
	if userRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	if err := s.products.Update(ctx, p); err != nil {
		return nil, err
	}

	_ = s.cache.Del(ctx, "products:id:"+p.ID)
	_ = s.cache.DelByPattern(ctx, "products:cat:*")
	return p, nil
}

// Delete elimina un producto. Solo admins.
func (s *ProductService) Delete(ctx context.Context, userRole, id string) error {
	if userRole != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	if err := s.products.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.cache.Del(ctx, "products:id:"+id)
	_ = s.cache.DelByPattern(ctx, "products:cat:*")
	return nil
}
