package ports

import (
	"context"

	"restaurants-e2/internal/domain"
)

// SearchIndex abstrae el backend de búsqueda (ElasticSearch en la implementación actual).
// El microservicio de búsqueda (cmd/search/) consume esta interfaz; si mañana se reemplaza
// ElasticSearch por OpenSearch o Meilisearch, solo cambia el adapter.
type SearchIndex interface {
	// IndexProduct indexa (o reindexa) un solo producto.
	// Si el producto no tiene descripción, el adapter aplica "Producto sin descripción".
	IndexProduct(ctx context.Context, p *domain.Product) error

	// BulkIndexProducts indexa muchos productos de una — usado por POST /search/reindex.
	BulkIndexProducts(ctx context.Context, ps []domain.Product) error

	// SearchProducts busca productos por texto libre sobre name/description/category.
	SearchProducts(ctx context.Context, query string, limit int) ([]domain.Product, error)

	// SearchByCategory filtra por categoría exacta.
	SearchByCategory(ctx context.Context, category string, limit int) ([]domain.Product, error)

	// DeleteProduct remueve un producto del índice (al eliminar en la BD).
	DeleteProduct(ctx context.Context, id string) error
}
