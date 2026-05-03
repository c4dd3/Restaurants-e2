package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LLMClient es la abstracción mínima para cualquier proveedor de LLM.
type LLMClient interface {
	Complete(ctx context.Context, system, user string) (string, error)
}

// AnthropicClient llama a la API de Anthropic (Claude Haiku).
type AnthropicClient struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *AnthropicClient) Complete(ctx context.Context, system, user string) (string, error) {
	body := map[string]any{
		"model":      c.model,
		"max_tokens": 4096,
		"system":     system,
		"messages": []map[string]string{
			{"role": "user", "content": user},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("respuesta vacía del LLM")
	}

	// Limpiar posibles bloques markdown ```json ... ```
	text := result.Content[0].Text
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text), nil
}
