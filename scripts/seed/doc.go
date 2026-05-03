// Package main implementa el generador de datos de prueba asistido por LLM.
//
// Uso:
//
//	go run ./scripts/seed [flags]
//	go run ./scripts/seed --restaurants=20 --menus-per=3 --products-per=10
//	go run ./scripts/seed --dry-run   # genera sin insertar en BD
//
// Variables de entorno requeridas:
//
//	ANTHROPIC_API_KEY  — API key de Anthropic
//	DB_ENGINE          — postgres | mongo
//	SEED_LLM_MODEL     — modelo a usar (default: claude-haiku-4-5-20251001)
package main
