# ============================================================================
# Restaurants-e2 — comandos comunes
#
# Uso: make <target>
# Ejemplos:
#   make up              # levanta todo con el engine por defecto (.env)
#   make up-postgres     # levanta solo con Postgres
#   make up-mongo        # levanta solo con Mongo (con sharding)
#   make scale-api N=3   # escala api a N réplicas
# ============================================================================

COMPOSE := docker compose -f deployments/docker-compose.yml

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ---------- Desarrollo local (sin Docker) ----------
.PHONY: run-api run-auth run-search
run-api:    ## Corre el servicio api directamente (requiere .env y dependencias Postgres/Redis vivas)
	go run ./cmd/api
run-auth:   ## Corre el servicio auth directamente
	go run ./cmd/auth
run-search: ## Corre el servicio search directamente
	go run ./cmd/search

.PHONY: tidy
tidy: ## Actualiza go.sum y descarga deps
	go mod tidy

.PHONY: build
build: ## Compila los 3 binarios en ./bin/
	mkdir -p bin
	go build -o bin/api    ./cmd/api
	go build -o bin/auth   ./cmd/auth
	go build -o bin/search ./cmd/search

# Tests: pospuestos a fase final del proyecto (ver README).

# ---------- Docker / Compose ----------
.PHONY: up up-postgres up-mongo down logs ps
up: ## docker compose up con el perfil por defecto
	$(COMPOSE) up --build -d

up-postgres: ## Levanta stack con Postgres
	DB_ENGINE=postgres $(COMPOSE) --profile postgres up --build -d

up-mongo: ## Levanta stack con Mongo (replica set + sharding)
	DB_ENGINE=mongo $(COMPOSE) --profile mongo up --build -d

down: ## Apaga y elimina contenedores
	$(COMPOSE) down

down-v: ## Apaga y borra volúmenes (pierde datos)
	$(COMPOSE) down -v

logs: ## Sigue logs de todos los servicios
	$(COMPOSE) logs -f

ps: ## Estado de los servicios
	$(COMPOSE) ps

# ---------- Escalado ----------
N ?= 2
.PHONY: scale-api scale-search
scale-api:    ## Escala api a N réplicas (make scale-api N=3)
	$(COMPOSE) up -d --scale api=$(N) --no-recreate
scale-search: ## Escala search a N réplicas
	$(COMPOSE) up -d --scale search=$(N) --no-recreate

# ---------- Utilidades ----------
.PHONY: seed health
seed: ## Corre el script de generación de datos con LLM
	go run ./scripts/seed

health: ## Chequea /health de todos los servicios vía el balanceador
	@echo "--- nginx ---"   && curl -s http://localhost/healthz      | head -1
	@echo "--- api ---"     && curl -s http://localhost/api/health   | head -1
	@echo "--- auth ---"    && curl -s http://localhost/auth/health  | head -1
	@echo "--- search ---"  && curl -s http://localhost/search/health| head -1

.PHONY: fmt vet
fmt: ## gofmt + goimports
	gofmt -s -w .
vet: ## go vet
	go vet ./...
