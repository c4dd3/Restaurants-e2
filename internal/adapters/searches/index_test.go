package searches

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"

	"restaurants-e2/internal/domain"
)

// Simulo Elastic con httptest para probar creación del índice, indexación y búsquedas
// sin pegarle al contenedor real.

func markAsElasticsearch(w http.ResponseWriter) {
	// El cliente oficial de elastic v8 revisa este header.
	// Si no está, asume que no está hablando con Elasticsearch de verdad.
	w.Header().Set("X-Elastic-Product", "Elasticsearch")
}

func TestIndexEnsureIndexAndSearch(t *testing.T) {
	var createdIndex bool
	var indexedBody string // me sirve para revisar qué terminó mandando el indexador

	// Este server mock responde solo lo que el índice necesita en estas rutas.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		markAsElasticsearch(w)

		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/products_test":
			if createdIndex {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

		case r.Method == http.MethodPut && r.URL.Path == "/products_test":
			createdIndex = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"acknowledged":true}`))

		case strings.Contains(r.URL.Path, "/_doc/prod-1") && (r.Method == http.MethodPut || r.Method == http.MethodPost):
			body, _ := io.ReadAll(r.Body)
			indexedBody = string(body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":"created"}`))

		case strings.Contains(r.URL.Path, "/_bulk"):
			body, _ := io.ReadAll(r.Body)
			indexedBody = string(body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errors":false,"items":[]}`))

		case strings.Contains(r.URL.Path, "/_search"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"hits":{"hits":[{"_source":{"id":"prod-1","menu_id":"menu-1","restaurant_id":"rest-1","name":"Pizza","description":"Producto sin descripción","category":"pizzas","price":4500,"available":true}}]}}`))

		case strings.Contains(r.URL.Path, "/_doc/prod-1") && r.Method == http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":"deleted"}`))

		default:
			t.Fatalf("request inesperado: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{server.URL},
	})
	if err != nil {
		t.Fatal(err)
	}

	idx, err := NewIndex(context.Background(), es, "products_test")
	if err != nil {
		t.Fatal(err)
	}

	if !createdIndex {
		t.Fatal("EnsureIndex no creó el índice")
	}

	err = idx.IndexProduct(context.Background(), &domain.Product{
		ID:           "prod-1",
		MenuID:       "menu-1",
		RestaurantID: "rest-1",
		Name:         "Pizza",
		Category:     "pizzas",
		Price:        4500,
		Available:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(indexedBody, domain.DefaultProductDescription) {
		t.Fatalf("no aplicó descripción por defecto: %s", indexedBody)
	}

	products, err := idx.SearchProducts(context.Background(), "pizza", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 || products[0].Name != "Pizza" {
		t.Fatalf("resultado de búsqueda incorrecto: %#v", products)
	}

	products, err = idx.SearchByCategory(context.Background(), "pizzas", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 {
		t.Fatalf("resultado por categoría incorrecto: %#v", products)
	}

	if err := idx.DeleteProduct(context.Background(), "prod-1"); err != nil {
		t.Fatal(err)
	}
}

// Si un producto viene sin ID, prefiero que no se meta al bulk.
func TestBulkIndexProductsOmiteProductosSinID(t *testing.T) {
	var bulkLines int

	// Este server mock responde solo lo que el índice necesita para este caso.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		markAsElasticsearch(w)

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}

		if !strings.Contains(r.URL.Path, "/_bulk") {
			t.Fatalf("request inesperado: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
			if line != "" {
				bulkLines++
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": false,
			"items":  []any{},
		})
	}))
	defer server.Close()

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{server.URL},
	})
	if err != nil {
		t.Fatal(err)
	}

	idx, err := NewIndex(context.Background(), es, "products_test")
	if err != nil {
		t.Fatal(err)
	}

	err = idx.BulkIndexProducts(context.Background(), []domain.Product{
		{ID: "prod-1", Name: "Pizza", Category: "pizzas"},
		{Name: "Sin ID"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if bulkLines != 2 {
		t.Fatalf("se esperaban 2 líneas bulk para 1 producto válido, obtuvo %d", bulkLines)
	}
}
