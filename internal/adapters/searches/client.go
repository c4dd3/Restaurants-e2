package searches

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"

	"restaurants-e2/internal/config"
)

func NewClient(cfg config.SearchConfig) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.URL},
	})
	if err != nil {
		return nil, err
	}

	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch no disponible: %s", res.Status())
	}

	return es, nil
}
