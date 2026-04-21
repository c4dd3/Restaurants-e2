-- ============================================================================
-- Schema inicial para PostgreSQL — Restaurants-e2
--
-- Notas de diseño:
--   - UUIDs como PK (gen_random_uuid() de pgcrypto) para portabilidad con Mongo.
--   - TIMESTAMPTZ siempre: evita el dolor de las zonas horarias.
--   - Índices mínimos pero relevantes para los queries esperados.
--   - ON DELETE CASCADE solo donde el borrado lógico no es razonable.
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- -------------------- Users --------------------
CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,             -- hash bcrypt
    role       VARCHAR(20)  NOT NULL CHECK (role IN ('client', 'admin')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- -------------------- Restaurants --------------------
CREATE TABLE IF NOT EXISTS restaurants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    address     VARCHAR(500) NOT NULL,
    phone       VARCHAR(50)  NOT NULL,
    description TEXT,
    admin_id    UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    capacity    INT NOT NULL CHECK (capacity > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_restaurants_admin ON restaurants(admin_id);

-- -------------------- Menus --------------------
CREATE TABLE IF NOT EXISTS menus (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_menus_restaurant ON menus(restaurant_id);

-- -------------------- Products (antes MenuItem) --------------------
CREATE TABLE IF NOT EXISTS products (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_id       UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    category      VARCHAR(100) NOT NULL,
    price         NUMERIC(10,2) NOT NULL CHECK (price > 0),
    available     BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_products_menu     ON products(menu_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
-- Para ElasticSearch-like búsqueda simple sobre name; si se requiere full-text agregamos tsvector después.
CREATE INDEX IF NOT EXISTS idx_products_name_lc  ON products(LOWER(name));

-- -------------------- Reservations --------------------
CREATE TABLE IF NOT EXISTS reservations (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id)       ON DELETE CASCADE,
    date          TIMESTAMPTZ NOT NULL,
    party_size    INT NOT NULL CHECK (party_size > 0),
    status        VARCHAR(20) NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending','confirmed','cancelled')),
    notes         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reservations_restaurant_date ON reservations(restaurant_id, date);
CREATE INDEX IF NOT EXISTS idx_reservations_user            ON reservations(user_id);

-- -------------------- Orders --------------------
CREATE TABLE IF NOT EXISTS orders (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id)           ON DELETE RESTRICT,
    restaurant_id  UUID NOT NULL REFERENCES restaurants(id)     ON DELETE RESTRICT,
    reservation_id UUID REFERENCES reservations(id)             ON DELETE SET NULL,
    total          NUMERIC(10,2) NOT NULL DEFAULT 0,
    status         VARCHAR(20) NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending','confirmed','cancelled')),
    pickup         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_user       ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_restaurant ON orders(restaurant_id);

CREATE TABLE IF NOT EXISTS order_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID NOT NULL REFERENCES orders(id)   ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INT NOT NULL CHECK (quantity > 0),
    price      NUMERIC(10,2) NOT NULL               -- snapshot del precio al momento de ordenar
);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
