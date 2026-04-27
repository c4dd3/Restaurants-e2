package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"restaurants-e2/internal/ports"
)

type SearchHandler struct {
	idx      ports.SearchIndex
	products ports.ProductRepository
}

func NewSearchHandler(idx ports.SearchIndex, products ports.ProductRepository) *SearchHandler {
	return &SearchHandler{idx: idx, products: products}
}

func (h *SearchHandler) RegisterRoutes(r gin.IRouter) {
	search := r.Group("/search")
	search.GET("/products", h.SearchProducts)
	search.GET("/products/category/:categoria", h.SearchProductsByCategory)
	search.POST("/reindex", h.Reindex)
}

func (h *SearchHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el parámetro q es obligatorio"})
		return
	}

	limit := parseLimit(c.DefaultQuery("limit", "20"))

	items, err := h.idx.SearchProducts(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query": query,
		"count": len(items),
		"items": items,
	})
}

func (h *SearchHandler) SearchProductsByCategory(c *gin.Context) {
	category := c.Param("categoria")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "la categoría es obligatoria"})
		return
	}

	limit := parseLimit(c.DefaultQuery("limit", "20"))

	items, err := h.idx.SearchByCategory(c.Request.Context(), category, limit)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": category,
		"count":    len(items),
		"items":    items,
	})
}

func (h *SearchHandler) Reindex(c *gin.Context) {
	start := time.Now()

	products, err := h.products.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.idx.BulkIndexProducts(c.Request.Context(), products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"indexed":     len(products),
		"duration_ms": time.Since(start).Milliseconds(),
	})
}

func parseLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return 20
	}
	if limit > 50 {
		return 50
	}
	return limit
}
