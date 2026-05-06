package http

// search_handler_test.go — tests adicionales para cubrir las ramas de error
// y límites de parseLimit que no están en TestSearchHandlerRoutes (auth_handler_test.go).
//
// NOTA: mockSearchIndex ya está definido en auth_handler_test.go con campos:
//   items []domain.Product
//   err   error

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

// ── mockProductRepoSearch ─────────────────────────────────────────────────────
// Repo de productos con inyección de error en FindAll — para probar Reindex.
// Se llama distinto a mockProductRepo para evitar redeclaración.

type mockProductRepoSearch struct {
	products   []domain.Product
	findAllErr error
}

func (r *mockProductRepoSearch) FindAll(_ context.Context) ([]domain.Product, error) {
	return r.products, r.findAllErr
}
func (r *mockProductRepoSearch) FindByID(_ context.Context, _ string) (*domain.Product, error) {
	return nil, domain.ErrNotFound
}
func (r *mockProductRepoSearch) FindByIDs(_ context.Context, _ []string) ([]domain.Product, error) {
	return nil, nil
}
func (r *mockProductRepoSearch) FindByCategory(_ context.Context, _ string) ([]domain.Product, error) {
	return nil, nil
}
func (r *mockProductRepoSearch) Create(_ context.Context, _ *domain.Product) error { return nil }
func (r *mockProductRepoSearch) Update(_ context.Context, _ *domain.Product) error { return nil }
func (r *mockProductRepoSearch) Delete(_ context.Context, _ string) error           { return nil }

var _ ports.ProductRepository = (*mockProductRepoSearch)(nil)

// ── helpers ────────────────────────────────────────────────────────────────────

func newSearchTestRouter(idx *mockSearchIndex, prods ports.ProductRepository) *gin.Engine {
	setupGin()
	h := NewSearchHandler(idx, prods)
	r := gin.New()
	h.RegisterRoutes(r)
	return r
}

// ── SearchProducts error ───────────────────────────────────────────────────────

func TestSearchHandlerSearchProductsError(t *testing.T) {
	// SearchProducts devuelve error → 503 ServiceUnavailable
	idx := &mockSearchIndex{err: errors.New("elastic caído")}
	r := newSearchTestRouter(idx, &mockProductRepoSearch{})

	w := performJSON(r, http.MethodGet, "/search/products?q=pizza", nil)
	requireStatus(t, w, http.StatusServiceUnavailable)
}

// ── SearchByCategory error ─────────────────────────────────────────────────────

func TestSearchHandlerSearchByCategoryError(t *testing.T) {
	// SearchByCategory devuelve error → 503 ServiceUnavailable
	idx := &mockSearchIndex{err: errors.New("elastic caído")}
	r := newSearchTestRouter(idx, &mockProductRepoSearch{})

	w := performJSON(r, http.MethodGet, "/search/products/category/pizzas", nil)
	requireStatus(t, w, http.StatusServiceUnavailable)
}

// ── Reindex errors ─────────────────────────────────────────────────────────────

func TestSearchHandlerReindexFindAllError(t *testing.T) {
	// FindAll falla → 500 Internal Server Error
	prods := &mockProductRepoSearch{findAllErr: errors.New("db error")}
	r := newSearchTestRouter(&mockSearchIndex{}, prods)

	w := performJSON(r, http.MethodPost, "/search/reindex", nil)
	requireStatus(t, w, http.StatusInternalServerError)
}

func TestSearchHandlerReindexBulkError(t *testing.T) {
	// BulkIndexProducts falla → 500 Internal Server Error
	prods := &mockProductRepoSearch{
		products: []domain.Product{{ID: "p-1", Name: "Pizza", Category: "pizzas"}},
	}
	idx := &mockSearchIndex{err: errors.New("bulk failed")}
	r := newSearchTestRouter(idx, prods)

	w := performJSON(r, http.MethodPost, "/search/reindex", nil)
	requireStatus(t, w, http.StatusInternalServerError)
}

// ── parseLimit branches ────────────────────────────────────────────────────────

func TestSearchHandlerParseLimitInvalidString(t *testing.T) {
	// limit no numérico → parseLimit retorna 20 (rama err != nil)
	idx := &mockSearchIndex{items: []domain.Product{{ID: "p-1", Name: "Pizza"}}}
	r := newSearchTestRouter(idx, &mockProductRepoSearch{})

	w := performJSON(r, http.MethodGet, "/search/products?q=pizza&limit=abc", nil)
	requireStatus(t, w, http.StatusOK)
}

func TestSearchHandlerParseLimitNegative(t *testing.T) {
	// limit negativo → parseLimit retorna 20 (rama limit <= 0)
	idx := &mockSearchIndex{items: []domain.Product{{ID: "p-1", Name: "Pizza"}}}
	r := newSearchTestRouter(idx, &mockProductRepoSearch{})

	w := performJSON(r, http.MethodGet, "/search/products?q=pizza&limit=-5", nil)
	requireStatus(t, w, http.StatusOK)
}
