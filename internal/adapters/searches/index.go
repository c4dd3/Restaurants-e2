package searches

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

type Index struct {
	es        *elasticsearch.Client
	indexName string
}

var _ ports.SearchIndex = (*Index)(nil)

func NewIndex(ctx context.Context, es *elasticsearch.Client, indexName string) (*Index, error) {
	if strings.TrimSpace(indexName) == "" {
		indexName = "products"
	}
	idx := &Index{es: es, indexName: indexName}
	if err := idx.EnsureIndex(ctx); err != nil {
		return nil, err
	}
	return idx, nil
}

func (idx *Index) EnsureIndex(ctx context.Context) error {
	res, err := idx.es.Indices.Exists([]string{idx.indexName}, idx.es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}
	if res.StatusCode != 404 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("validar índice: %s %s", res.Status(), string(body))
	}

	mapping := `{
		"mappings": {
			"properties": {
				"id": { "type": "keyword" },
				"menu_id": { "type": "keyword" },
				"restaurant_id": { "type": "keyword" },
				"name": { "type": "text" },
				"description": { "type": "text" },
				"category": { "type": "keyword" },
				"price": { "type": "float" },
				"available": { "type": "boolean" }
			}
		}
	}`

	createRes, err := idx.es.Indices.Create(
		idx.indexName,
		idx.es.Indices.Create.WithContext(ctx),
		idx.es.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		body, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("crear índice: %s %s", createRes.Status(), string(body))
	}

	return nil
}

func (idx *Index) IndexProduct(ctx context.Context, p *domain.Product) error {
	if p == nil {
		return errors.New("producto nil")
	}
	copyProduct := *p
	if copyProduct.Description == "" {
		copyProduct.Description = domain.DefaultProductDescription
	}

	body, err := json.Marshal(copyProduct)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      idx.indexName,
		DocumentID: copyProduct.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, idx.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		responseBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("indexar producto: %s %s", res.Status(), string(responseBody))
	}

	return nil
}

func (idx *Index) BulkIndexProducts(ctx context.Context, products []domain.Product) error {
	if len(products) == 0 {
		return nil
	}

	var bulk bytes.Buffer
	enc := json.NewEncoder(&bulk)

	for _, product := range products {
		if product.ID == "" {
			continue
		}
		if product.Description == "" {
			product.Description = domain.DefaultProductDescription
		}

		meta := map[string]map[string]string{
			"index": {
				"_index": idx.indexName,
				"_id":    product.ID,
			},
		}
		if err := enc.Encode(meta); err != nil {
			return err
		}
		if err := enc.Encode(product); err != nil {
			return err
		}
	}

	if bulk.Len() == 0 {
		return nil
	}

	res, err := idx.es.Bulk(
		bytes.NewReader(bulk.Bytes()),
		idx.es.Bulk.WithContext(ctx),
		idx.es.Bulk.WithRefresh("wait_for"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.IsError() {
		return fmt.Errorf("bulk index: %s %s", res.Status(), string(body))
	}

	var parsed struct {
		Errors bool `json:"errors"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return err
	}
	if parsed.Errors {
		return fmt.Errorf("bulk index terminó con errores: %s", string(body))
	}

	return nil
}

func (idx *Index) SearchProducts(ctx context.Context, query string, limit int) ([]domain.Product, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	body := map[string]any{
		"size": limit,
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":     query,
				"fields":    []string{"name^3", "category^2", "description"},
				"fuzziness": "AUTO",
			},
		},
	}

	return idx.search(ctx, body)
}

func (idx *Index) SearchByCategory(ctx context.Context, category string, limit int) ([]domain.Product, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	body := map[string]any{
		"size": limit,
		"query": map[string]any{
			"term": map[string]any{
				"category": category,
			},
		},
	}

	return idx.search(ctx, body)
}

func (idx *Index) DeleteProduct(ctx context.Context, id string) error {
	res, err := idx.es.Delete(idx.indexName, id, idx.es.Delete.WithContext(ctx), idx.es.Delete.WithRefresh("wait_for"))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil
	}
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("eliminar producto del índice: %s %s", res.Status(), string(body))
	}

	return nil
}

func (idx *Index) search(ctx context.Context, body map[string]any) ([]domain.Product, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	res, err := idx.es.Search(
		idx.es.Search.WithContext(ctx),
		idx.es.Search.WithIndex(idx.indexName),
		idx.es.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseBody, _ := io.ReadAll(res.Body)
	if res.IsError() {
		return nil, fmt.Errorf("buscar productos: %s %s", res.Status(), string(responseBody))
	}

	var parsed struct {
		Hits struct {
			Hits []struct {
				Source domain.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return nil, err
	}

	products := make([]domain.Product, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		products = append(products, hit.Source)
	}

	return products, nil
}
