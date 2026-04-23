package main

// llm.go — cliente HTTP genérico para el proveedor LLM (OpenAI o Anthropic).
//
// Abstracción mínima: una interface LLMClient con un único método Complete
// que acepta un system prompt, un user prompt y devuelve un string (la
// respuesta, esperada como JSON).
//
//   type LLMClient interface {
//       Complete(ctx context.Context, system, user string) (string, error)
//   }
//
// Implementaciones:
//
//   OpenAIClient:
//     - Endpoint: https://api.openai.com/v1/chat/completions
//     - Header: Authorization: Bearer <OPENAI_API_KEY>
//     - Body: {
//         "model": "gpt-4o-mini",
//         "messages": [{"role":"system","content":system},{"role":"user","content":user}],
//         "response_format": {"type":"json_object"},
//         "temperature": 0.7
//       }
//     - Extraer choices[0].message.content.
//
//   AnthropicClient:
//     - Endpoint: https://api.anthropic.com/v1/messages
//     - Headers: x-api-key, anthropic-version: 2023-06-01.
//     - Body: {
//         "model": "claude-haiku-4-5",
//         "system": system,
//         "messages": [{"role":"user","content":user}],
//         "max_tokens": 4096
//       }
//     - Extraer content[0].text.
//
// Factory:
//
//   func NewLLMClient(cfg *config.Config) (LLMClient, error)
//     → decide según cfg.LLMProvider.
//
// Políticas:
//   - Timeout por request: 60s (los prompts largos pueden tardar).
//   - Retry: 3 intentos con backoff exponencial (2s, 4s, 8s) en 429/5xx.
//   - Rate limiting: cap de 1 req/s para no quemar la quota.
//   - Logging: loguear el prompt (truncado) y los primeros 200 chars de la
//     respuesta. No loguear la key.
//
// Parseo robusto:
//   - Antes de json.Unmarshal, limpiar la respuesta (algunos LLMs envuelven
//     JSON en ```json ... ``` aun si se pide response_format).
//   - Si el parse falla, re-prompt: "Por favor devuelve SOLO JSON válido,
//     sin markdown" y reintentar una vez.
//
// Costos estimados (para referencia del usuario):
//   - Haiku 4.5: ~$0.001 por restaurante+menú+productos. Seed completo ~$0.10.
//   - GPT-4o-mini: similar.
//   No usar Opus/GPT-4o porque es overkill para datos de seed.
