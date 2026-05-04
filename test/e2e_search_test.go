//go:build e2e

package test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestE2ESearchFlows(t *testing.T) {
	suffix := uniqueSuffix()
	admin := registerUser(t, "Admin Search", fmt.Sprintf("admin-search-%s@example.com", suffix), "secret123", "admin")
	restaurant := mustCreateRestaurant(t, admin.Token, suffix)
	menu := mustCreateMenu(t, admin.Token, restaurant.ID, suffix)

	targetName := menu.Products[0].Name
	targetCategory := menu.Products[0].Category

	// Reindex manual -> 200
	resp := doJSONRequest(t, http.MethodPost, "/search/reindex", nil, "")
	if resp.StatusCode != http.StatusOK {
		raw := readBody(t, resp)
		t.Fatalf("reindex debía dar 200 y dio %d body=%s", resp.StatusCode, string(raw))
	}

	// Búsqueda textual -> 200 y debe incluir el producto creado en este test.
	resp = doJSONRequest(t, http.MethodGet, "/search/products?q="+url.QueryEscape(targetName), nil, "")
	if resp.StatusCode != http.StatusOK {
		raw := readBody(t, resp)
		t.Fatalf("search text debía dar 200 y dio %d body=%s", resp.StatusCode, string(raw))
	}
	searchByText := decodeJSON[searchResponse](t, resp)
	if !containsProductByName(searchByText.Items, targetName) {
		t.Fatalf("search text no devolvió el producto esperado: %#v", searchByText)
	}

	// Búsqueda por categoría -> 200
	resp = doJSONRequest(t, http.MethodGet, "/search/products/category/"+url.PathEscape(targetCategory), nil, "")
	if resp.StatusCode != http.StatusOK {
		raw := readBody(t, resp)
		t.Fatalf("search category debía dar 200 y dio %d body=%s", resp.StatusCode, string(raw))
	}
	searchByCategory := decodeJSON[searchResponse](t, resp)
	if !containsProductByName(searchByCategory.Items, targetName) {
		t.Fatalf("search category no devolvió el producto esperado: %#v", searchByCategory)
	}

	// Producto sin descripción -> Elastic usa el texto por defecto al indexar.
	resp = doJSONRequest(t, http.MethodGet, "/search/products?q="+url.QueryEscape(menu.Products[1].Name), nil, "")
	if resp.StatusCode != http.StatusOK {
		raw := readBody(t, resp)
		t.Fatalf("search bebida debía dar 200 y dio %d body=%s", resp.StatusCode, string(raw))
	}
	raw := string(readBody(t, resp))
	if !strings.Contains(raw, "Producto sin descripción") {
		t.Fatalf("search no mostró descripción por defecto. body=%s", raw)
	}

	// q faltante -> 400
	resp = doJSONRequest(t, http.MethodGet, "/search/products", nil, "")
	requireStatus(t, resp, http.StatusBadRequest)
}
