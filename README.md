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
├── scripts/seed/     # Generador de datos realistas con LLM
├── test/             # Tests e2e (Postgres + Mongo)
├── .github/workflows/ci.yml
├── Dockerfile.api | Dockerfile.auth | Dockerfile.search
└── Makefile
```

## Requisitos previos

- Docker 24+ y Docker Compose v2
- Go 1.23+ (solo para correr tests localmente o fuera de contenedor)
- Make (opcional — hay atajos en el `Makefile`)

## Puesta en marcha

### 1. Configuración

```bash
cp .env.example .env
# Elegir DB_ENGINE=postgres o DB_ENGINE=mongo
# Cambiar JWT_SECRET por un valor random de 32+ caracteres
```

### 2. Modo Postgres (por defecto)

```bash
DB_ENGINE=postgres docker compose -f deployments/docker-compose.yml --profile postgres up --build -d
```

### 3. Modo MongoDB (con replica set + sharding)

```bash
DB_ENGINE=mongo docker compose -f deployments/docker-compose.yml --profile mongo up --build -d
```

El contenedor `mongo_init` inicializa automáticamente el cluster al levantar:

1. Replica set del config server (`cfgrs`)
2. Replica sets `shard1rs` y `shard2rs` — cada uno con 1 primario + 2 secundarios
3. Agrega los shards al cluster vía `mongos`
4. Habilita sharding en la BD `restaurants`
5. Define shard keys: `products.category` (hashed), `reservations.restaurant_id` (hashed)

Los servicios esperan a que `mongos` esté healthy y `mongo_init` haya terminado antes de arrancar. El stack completo tarda ~60-90 segundos en estar listo.

Para verificar el estado del cluster:
```bash
docker exec -it re2_mongos mongosh --eval 'sh.status()'
```

### 4. Cambio de motor sin recompilar

```bash
# Bajar el stack y levantarlo con el otro motor
docker compose -f deployments/docker-compose.yml --profile postgres down
DB_ENGINE=mongo docker compose -f deployments/docker-compose.yml --profile mongo up --build -d
# No se cambia una sola línea de Go.
```

### 5. Verificar que todo responde

```bash
curl -s http://localhost/healthz
curl -s http://localhost/api/health
curl -s http://localhost/auth/health
curl -s http://localhost/search/health
```

### 6. Escalar horizontalmente

Los servicios `api`, `auth` y `search` se pueden escalar de forma independiente. Para levantar el stack con un número de réplicas específico desde el inicio:

```bash
# Postgres — réplicas iguales
DB_ENGINE=postgres docker compose -f deployments/docker-compose.yml --profile postgres up --build -d \
  --scale api=3 --scale auth=3 --scale search=3

# Postgres — réplicas por separado
DB_ENGINE=postgres docker compose -f deployments/docker-compose.yml --profile postgres up --build -d \
  --scale api=3 --scale auth=2 --scale search=2

# Mongo — mismo esquema
DB_ENGINE=mongo docker compose -f deployments/docker-compose.yml --profile mongo up --build -d \
  --scale api=3 --scale auth=2 --scale search=2
```

Para cambiar réplicas en caliente sin bajar el stack:

```bash
# Escala api a 5 sin tocar auth ni search
docker compose -f deployments/docker-compose.yml --profile postgres up -d --scale api=5 --no-recreate
```

Si una instancia muere o se mata manualmente, para volver al número deseado:

```bash
# Compose levanta las que faltan para llegar a 3, sin tocar las vivas
docker compose -f deployments/docker-compose.yml --profile postgres up -d --scale api=3 --no-recreate
```

Nginx descubre las instancias automáticamente vía el resolver DNS interno de Docker (`127.0.0.11`, re-resolución cada 5 s). Si una instancia muere mid-request, `proxy_next_upstream` reintenta en otra instancia viva sin que el cliente perciba el error.

Para verificar el balanceo y el failover:

```bash
# Ver a qué instancia va cada request
docker logs re2_nginx 2>&1 | grep 'upstream=' | tail -10

# Matar una instancia y verificar que el sistema sigue respondiendo
docker kill deployments-api-1
for i in $(seq 1 5); do curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost/api/restaurants; done
```

> **Nota sobre Kubernetes**: No se utiliza Kubernetes como opción de escalado horizontal. Se optó por Docker Compose con `deploy.replicas` dado que el despliegue final es en entorno local, lo que evita la sobrecarga operativa sin sacrificar la demostración del concepto. En producción, los mismos Dockerfiles se desplegarían como Deployments de Kubernetes sin modificar el código de la aplicación.

## Endpoints

### A través del balanceador (`http://localhost`)

| Método | Ruta | Auth | Descripción |
|--------|------|------|-------------|
| POST | `/auth/register` | — | Registro de usuario (`role: client\|admin`) |
| POST | `/auth/login` | — | Login, devuelve JWT |
| GET | `/api/users/me` | JWT | Perfil del usuario autenticado |
| PUT | `/api/users/:id` | JWT | Actualizar nombre o email |
| DELETE | `/api/users/:id` | JWT | Eliminar cuenta |
| GET | `/api/restaurants` | — | Listar restaurantes |
| GET | `/api/restaurants/:id` | — | Ver restaurante por ID |
| POST | `/api/restaurants` | Admin | Crear restaurante |
| POST | `/api/menus` | Admin | Crear menú con productos |
| GET | `/api/menus/:id` | JWT | Ver menú |
| PUT | `/api/menus/:id` | Admin | Actualizar menú |
| DELETE | `/api/menus/:id` | Admin | Eliminar menú |
| GET | `/api/products?category=X` | JWT | Listar productos por categoría |
| GET | `/api/products/:id` | JWT | Ver producto |
| PATCH | `/api/products/:id` | Admin | Actualizar producto |
| DELETE | `/api/products/:id` | Admin | Eliminar producto |
| POST | `/api/reservations` | JWT | Crear reserva |
| DELETE | `/api/reservations/:id` | JWT | Cancelar reserva |
| POST | `/api/orders` | JWT | Crear pedido |
| GET | `/api/orders/:id` | JWT | Ver pedido |
| GET | `/search/products?q=texto` | — | Búsqueda textual en productos |
| GET | `/search/products/category/:cat` | — | Filtrar por categoría |
| POST | `/search/reindex` | Admin | Reindexar productos manualmente |
| GET | `/healthz` | — | Health del balanceador Nginx |

Todos los errores tienen formato uniforme: `{"error": "código_estable", "detail": "..."}`.

## Redis — caché y políticas de expiración

Redis se usa como capa de caché **cache-aside** en los servicios de productos, menús y restaurantes. El flujo es: intentar leer de caché → si miss, leer de BD y cachear → en escrituras/borrados, invalidar las claves afectadas.

| Recurso | Clave Redis | TTL |
|---------|-------------|-----|
| Producto por ID | `products:id:<id>` | 10 min |
| Productos por categoría | `products:cat:<cat>` | 10 min |
| Restaurante por ID | `restaurants:id:<id>` | 5 min |
| Listado de restaurantes | `restaurants:all` | 5 min |
| Menú por ID | `menus:id:<id>` | 5 min |
| Disponibilidad de reservas | `reservations:rest:<id>:<fecha>` | invalidación por escritura |

**Política de evicción**: Redis corre con `maxmemory 256mb` y `maxmemory-policy allkeys-lru`. Cuando el caché llena los 256 MB, evicta automáticamente las claves menos recientemente usadas sin gestión manual.

**Invalidación activa**: al actualizar o borrar un recurso, el servicio llama a `cache.Del` o `cache.DelByPattern` para eliminar las claves afectadas. La búsqueda por patrón usa `SCAN` en lotes de 500 (nunca `KEYS`) para no bloquear Redis en producción.

## Testing

Las pruebas viven en el branch `tests` e incluyen tres niveles:

- **Unitarias** (`internal/service/`, `internal/auth/`, `internal/domain/`): validan la lógica de negocio con mocks de repositorios.
- **Integración** (`internal/adapters/repopg/`, `internal/adapters/repomongo/`): validan los adapters contra una BD real (Postgres y MongoDB).
- **End-to-end** (`test/`, `internal/e2e/`): flujos completos registrar → login → crear restaurante → reservar → ordenar.

Para correr los tests localmente (requiere Postgres en `:5432` y Mongo en `:27017`):

```bash
POSTGRES_TEST_URL=postgres://postgres:postgres@localhost:5432/restaurants \
MONGO_TEST_URI=mongodb://localhost:27017 \
go test ./internal/... -count=1 -coverprofile=coverage.out -timeout 120s

# Ver cobertura total
go tool cover -func=coverage.out | tail -1
```

El CI verifica automáticamente que la cobertura sea ≥ 90% en cada push al branch `tests`.

## CI/CD

El pipeline vive en `.github/workflows/ci.yml` y se ejecuta en cada push:

| Branch | Jobs que corren |
|--------|----------------|
| `tests` | lint → tests (cobertura ≥ 90%) |
| `dev` | lint → build de las 3 imágenes Docker |
| `main` | lint → build → push a GHCR |

1. **lint**: `go vet` + `staticcheck` sobre todo el árbol.
2. **test** (solo `tests`): pruebas unitarias e integración contra Postgres 16 y MongoDB 7 levantados como services de GitHub Actions.
3. **build**: construye las imágenes `api`, `auth` y `search` en paralelo con matrix strategy y caché de capas.
4. **push** (solo `main`): publica las imágenes a `ghcr.io/<owner>/<repo>/<servicio>:latest` y `:<sha>`.

## Datos de prueba con LLM

El seeder usa la API de Anthropic para generar datos realistas (restaurantes, menús, productos y usuarios) y los inserta en la BD.

```bash
# Requiere ANTHROPIC_API_KEY en .env y el stack levantado
make seed

# Ver qué generaría sin insertar nada
go run ./scripts/seed --dry-run

# Configurar volumen de datos
go run ./scripts/seed -restaurants=10 -menus-per=2 -products-per=8 -users=20
```

## Documentación adicional

La carpeta `docs/` contiene ocho diagramas interactivos. Para abrirlos:

```bash
open docs/index.html
# o servir como HTTP: python3 -m http.server 8000 --directory docs
```

| # | Diagrama |
|---|----------|
| 1 | Topología de servicios |
| 2 | Hexagonal + DAO |
| 3 | Dependencias de paquetes Go |
| 4 | El único `if` del motor (pg vs mongo) |
| 5 | Ciclo de vida de un request (`POST /reservations`) |
| 6 | Topología de despliegue Docker |
| 7 | Modelo de datos (pg y mongo) |
| 8 | Contrato de API completo |

---

## Créditos

Proyecto universitario — Tecnológico de Costa Rica, Base de Datos 2, Prof. Kenneth Obando Rodríguez.
