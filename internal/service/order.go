package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// OrderService gestiona pedidos de comida.
// El total SIEMPRE se calcula en el servidor con los precios actuales de BD
// — nunca se acepta el total que envía el cliente.
type OrderService struct {
	orders      ports.OrderRepository
	products    ports.ProductRepository
	restaurants ports.RestaurantRepository
}

// NewOrderService construye el servicio inyectando sus dependencias.
func NewOrderService(
	orders ports.OrderRepository,
	products ports.ProductRepository,
	restaurants ports.RestaurantRepository,
) *OrderService {
	return &OrderService{
		orders:      orders,
		products:    products,
		restaurants: restaurants,
	}
}

// Create valida el pedido, calcula el total y lo persiste.
// El adapter envuelve order + order_items en una sola transacción.
func (s *OrderService) Create(ctx context.Context, userID string, req domain.CreateOrderRequest) (*domain.Order, error) {
	// 1. Validar que el restaurante exista.
	if _, err := s.restaurants.FindByID(ctx, req.RestaurantID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrValidation
		}
		return nil, err
	}

	// 2. Obtener todos los productos en una sola query (evita N+1).
	productIDs := make([]string, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	productList, err := s.products.FindByIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	// Indexar por ID para lookup O(1).
	productMap := make(map[string]*domain.Product, len(productList))
	for i := range productList {
		productMap[productList[i].ID] = &productList[i]
	}

	// 3. Validar existencia y pertenencia al restaurante; calcular total.
	items := make([]domain.OrderItem, 0, len(req.Items))
	var total float64

	for _, item := range req.Items {
		p, ok := productMap[item.ProductID]
		if !ok {
			return nil, fmt.Errorf("%w: producto %s no encontrado", domain.ErrValidation, item.ProductID)
		}
		if p.RestaurantID != req.RestaurantID {
			return nil, fmt.Errorf("%w: producto %s no pertenece al restaurante", domain.ErrValidation, item.ProductID)
		}

		// Precio del servidor — nunca el del cliente.
		total += p.Price * float64(item.Quantity)

		items = append(items, domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     p.Price, // precio congelado al momento del pedido
		})
	}

	// 4. Construir y persistir la orden.
	o := &domain.Order{
		ID:            uuid.New().String(),
		UserID:        userID,
		RestaurantID:  req.RestaurantID,
		ReservationID: req.ReservationID,
		Items:         items,
		Total:         total,
		Status:        domain.StatusPending,
		Pickup:        req.Pickup,
	}

	if err := s.orders.Create(ctx, o); err != nil {
		return nil, err
	}

	return o, nil
}

// GetByID devuelve un pedido con sus items. Solo el propietario o un admin puede verlo.
func (s *OrderService) GetByID(ctx context.Context, userID, role, orderID string) (*domain.Order, error) {
	o, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if o.UserID != userID && role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}

	return o, nil
}
