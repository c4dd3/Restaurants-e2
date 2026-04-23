# Restaurants-e2

API distribuida para reservas y pedidos en restaurantes. Etapa 2 del Proyecto 1 del curso Base de Datos 2 (Tecnológico de Costa Rica). Extiende la [Etapa 1](https://github.com/c4dd3/restaurant-api-bd2) con arquitectura de microservicios, persistencia dual Postgres/MongoDB, búsqueda con ElasticSearch, caché Redis, balanceador Nginx y CI/CD.

## Arquitectura

```
                           ┌───────────────┐
                     HTTP  │    Nginx      │   puerto 80
             ┌─────────────┤  load balancer├─────────────┐
             │             └───────┬───────┘             │
             │                     │                     │
     /auth/* │              /api/* │              /search/*
             ▼                     ▼                     ▼
      ┌──────────┐          ┌──────────┐          ┌──────────┐
      │   auth   │          │   api    │──────┐   │  search  │
      │  :8081   │          │  :8080   │      │   │  :8082   │
      └─────┬────┘          └────┬─────┘      │   └────┬─────┘
            │                    │            │        │
            │   JWT_SECRET       │            ▼        │
            │   compartido       │      ┌─────────┐    │
            └────────────────────┘      │  Redis  │    │
                                 │      │  cache  │    │
                 ┌───────────────┤      └─────────┘    │
                 │ DB_ENGINE=... │                     │
                 ▼               ▼                     ▼
          ┌───────────┐   ┌──────────────┐      ┌──────────────┐
          │ Postgres  │   │   MongoDB    │      │ElasticSearch │
          │ 16-alpine │   │  (sharded)   │      │     8.x      │
          └───────────┘   │  cfgrs +     │      └──────────────┘
                          │  2 shards    │
                          │  3 nodos c/u │
                          └──────────────┘
```

El api-service y el auth-service comparten `JWT_SECRET` para validar tokens de forma stateless (sin llamadas HTTP entre ellos). El search-service consume productos del repositorio principal para reindexar, pero sirve sus consultas desde ElasticSearch.

## Principios arquitectónicos

- **Hexagonal / Ports-and-Adapters**: `internal/ports/` define las interfaces que `internal/service/` consume; las implementaciones viven en `internal/adapters/`. Esto permite cambiar entre Postgres y Mongo vía variable de entorno (`DB_ENGINE=postgres|mongo`) sin tocar lógica de negocio.
- **Domain purity**: `internal/domain/` no importa nada fuera de la librería estándar. Cada struct tiene tags `json`, `db` y `bson` simultáneamente para poder persistirse en cualquiera de los dos motores.
- **Shutdown ordenado**: cada `main.go` escucha SIGTERM y termina conexiones en vuelo antes de morir — crítico con escalado horizontal.

## Estructura del repositorio

```
restaurants-e2/
├── cmd/
│   ├── api/          # Servicio principal (CRUD)
│   ├── auth/         # Servicio de autenticación (JWT)
│   └── search/       # Servicio de búsqueda (ElasticSearch)
├── internal/
│   ├── domain/       # Entidades puras del negocio
│   ├── ports/        # Interfaces (Repository, Cache, SearchIndex)
│   ├── service/      # Lógica de negocio (agnóstica de BD)
│   ├── adapters/
│   │   ├── repopg/        # Repositorios Postgres (pgx/v5)
│   │   ├── repomongo/     # Repositorios MongoDB (mongo-go-driver)
│   │   ├── cacheredis/    # Cliente Redis
│   │   └── searches/      # Cliente ElasticSearch
│   ├── transport/http/   # Handlers Gin + middleware
│   └── config/       # Carga de variables de entorno
├── deployments/
│   ├── docker-compose.yml
│   ├── nginx/nginx.conf
│   ├── mongo/init-cluster.sh
│   └── postgres/init.sql
├── scripts/seed/     # Generador de datos realistas
├── .github/workflows/ci.yml
├── Dockerfile.api | Dockerfile.auth | Dockerfile.search
└── Makefile
```

## Requisitos previos

- Docker 24+ y Docker Compose v2
- Go 1.22+ (solo si vas a correr sin contenedor o correr tests localmente)
- Make (opcional — hay atajos en el `Makefile`)

## Puesta en marcha

### 1. Configuración

```bash
cp .env.example .env
# Editar .env y elegir DB_ENGINE=postgres o DB_ENGINE=mongo
# Cambiar JWT_SECRET por un valor random de 32+ caracteres.
```

### 2. Modo Postgres (por defecto)

```bash
make up-postgres
# o: docker compose -f deployments/docker-compose.yml --profile postgres up --build -d
```

### 3. Modo MongoDB (con replica set + sharding)

```bash
make up-mongo
# El contenedor mongo_init se encarga de:
#   1. Iniciar el replica set del config server (cfgrs)
#   2. Iniciar los replica sets shard1rs y shard2rs (1 primario + 2 secundarios cada uno)
#   3. Agregar los shards al cluster vía mongos
#   4. Habilitar sharding en la BD `restaurants`
#   5. Definir shard keys: products.category (hashed), reservations.restaurant_id (hashed)
# Podés verificar con:
docker exec -it re2_mongos mongosh --eval 'sh.status()'
```

### 4. Cambio de motor sin recompilar

```bash
# Editar .env → DB_ENGINE=mongo
docker compose -f deployments/docker-compose.yml restart api auth search
# Listo. No se cambia una sola línea de Go.
```

### 5. Escalar horizontalmente

```bash
make scale-api N=3
# Nginx detecta las nuevas réplicas por DNS interno de Docker
# y empieza a balancear entre ellas automáticamente.
```

## Endpoints

### A través del balanceador (`http://localhost`)

| Método | Ruta                                   | Destino          | Descripción                       |
| ------ | -------------------------------------- | ---------------- | --------------------------------- |
| POST   | `/auth/register`                       | auth             | Registro de usuario               |
| POST   | `/auth/login`                          | auth             | Login, devuelve JWT               |
| GET    | `/api/users/me`                        | api              | Perfil del usuario autenticado    |
| GET    | `/api/restaurants`                     | api              | Listar restaurantes               |
| POST   | `/api/restaurants`                     | api (admin)      | Crear restaurante                 |
| POST   | `/api/menus`                           | api (admin)      | Crear menú con productos          |
| POST   | `/api/reservations`                    | api              | Crear reserva                     |
| POST   | `/api/orders`                          | api              | Crear pedido                      |
| GET    | `/search/products?q=texto`             | search           | Búsqueda textual                  |
| GET    | `/search/products/category/:categoria` | search           | Filtro por categoría              |
| POST   | `/search/reindex`                      | search (admin)   | Reindexar productos manualmente   |

## Testing

Pospuesto a la fase final. Cuando se agregue, la suite correrá contra los mismos ports que consume `service/`, de forma que pg-adapter y mongo-adapter se validen con los mismos casos.

## CI/CD

El pipeline vive en `.github/workflows/ci.yml` y en cada push:

1. Corre `go vet` + staticcheck sobre todo el árbol.
2. Construye las 3 imágenes Docker en paralelo (matrix strategy).
3. Solo en push a `main`: publica las imágenes a `ghcr.io/<owner>/<repo>/<servicio>:latest` y `:<sha>`.

## Datos de prueba (LLM)

```bash
make seed
# Usa un LLM para generar productos, menús y descripciones realistas.
# Ver scripts/seed/ para configuración.
```

## Documentación adicional

Abrí `docs/index.html` en el navegador (o serví la carpeta con `python -m http.server` desde `docs/`). Navega seis diagramas:

1. **Topología de servicios** — cómo se relacionan los 3 binarios con Nginx al frente.
2. **Hexagonal + DAO** — núcleo, ports y adapters con zoom a un DAO concreto.
3. **Dependencias de paquetes** — qué puede importar qué en Go.
4. **El único `if` del motor** — cómo `wiring.NewRepositories` concentra la decisión pg/mongo.
5. **Ciclo de vida de un request** — `POST /reservations` paso a paso.
6. **Topología de despliegue** — contenedores, red y puertos del compose.

---

## Créditos

Proyecto universitario — Tecnológico de Costa Rica, Base de Datos 2, Prof. Kenneth Obando Rodríguez.
