package service

// OrderService — casos de uso sobre pedidos.
//
// Un pedido consiste en una lista de items (producto + cantidad). El total se
// CALCULA en el servidor con los precios vigentes — nunca se confía en el
// total que manda el cliente.
//
// Dependencias:
//   - ports.OrderRepository
//   - ports.ProductRepository       (validar productos + obtener precios)
//   - ports.RestaurantRepository    (validar que el restaurante exista)
//
// Métodos públicos:
//
//   Create(ctx, userID string, req CreateOrderRequest) (*domain.Order, error)
//     1. Validar que el restaurante exista.
//     2. Obtener los productos en UNA sola query:
//        - ProductRepository.FindByIDs(productIDs)   // nuevo método, agregarlo al port si hace falta
//        - Si alguno no aparece en el resultado → ErrValidation (producto inexistente).
//     3. Validar que TODOS los productos pertenecen al menú del restaurante.
//        Si no → ErrValidation (producto no pertenece al restaurante).
//     4. Calcular total:
//        - total = Σ (productos[item.ProductID].Price × item.Quantity)
//        - Usar el precio del servidor, NO el del DTO.
//     5. Construir domain.Order (uuid + user_id + restaurant_id + items + total + status=pending).
//     6. OrderRepository.Create(ctx, &o)
//        - El adapter debe envolver la inserción de order + order_items en UNA transacción.
//          Si falla a mitad, ambas caen (rollback) → no quedan items huérfanos.
//     7. Devolver la orden.
//
//   GetByID(ctx, userID, orderID string) (*domain.Order, error)
//     1. OrderRepository.FindByID.
//     2. Si no existe → ErrNotFound.
//     3. Si order.UserID != userID AND role != admin → ErrForbidden.
//
// Observación pedagógica:
//   Este service es un buen ejemplo de por qué el PATRÓN hexagonal importa.
//   Necesitamos productos, restaurantes y órdenes — TRES sub-DAOs distintos —
//   pero al service le da igual si cada uno habla con pg o mongo. Solo ve
//   interfaces.
