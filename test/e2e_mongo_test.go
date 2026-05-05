//go:build e2e

package test

import (
    "fmt"
    "net/http"
    "strings"
    "testing"
    "time"
)

func TestE2EMongoAuthFlows(t *testing.T) {
    suffix := uniqueSuffix()
    email := fmt.Sprintf("auth-%s@example.com", suffix)
    password := "secret123"

    // Registro exitoso -> 201
    resp := doJSONRequest(t, http.MethodPost, "/auth/register", map[string]any{
        "name":     "Auth User",
        "email":    email,
        "password": password,
        "role":     "client",
    }, "")
    authOK := decodeJSON[authResponse](t, resp)
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("registro exitoso debía dar 201 y dio %d", resp.StatusCode)
    }
    if authOK.Token == "" {
        t.Fatal("registro exitoso no devolvió token")
    }

    // Registro duplicado -> 409
    resp = doJSONRequest(t, http.MethodPost, "/auth/register", map[string]any{
        "name":     "Auth User",
        "email":    email,
        "password": password,
        "role":     "client",
    }, "")
    requireStatus(t, resp, http.StatusConflict)

    // Login correcto -> 200 + token
    resp = doJSONRequest(t, http.MethodPost, "/auth/login", map[string]any{
        "email":    email,
        "password": password,
    }, "")
    loginOK := decodeJSON[authResponse](t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("login correcto debía dar 200 y dio %d", resp.StatusCode)
    }
    if loginOK.Token == "" {
        t.Fatal("login correcto no devolvió token")
    }

    // Login con password incorrecta -> 401
    resp = doJSONRequest(t, http.MethodPost, "/auth/login", map[string]any{
        "email":    email,
        "password": "mal-mal-mal",
    }, "")
    requireStatus(t, resp, http.StatusUnauthorized)

    // Ruta protegida sin token -> 401
    resp = doJSONRequest(t, http.MethodGet, "/api/users/me", nil, "")
    requireStatus(t, resp, http.StatusUnauthorized)

    // Ruta protegida con token inválido -> 401
    resp = doJSONRequest(t, http.MethodGet, "/api/users/me", nil, "esto-no-es-un-jwt")
    requireStatus(t, resp, http.StatusUnauthorized)
}

func TestE2EMongoRestaurantFlows(t *testing.T) {
    suffix := uniqueSuffix()
    admin := registerUser(t, "Admin Rest", fmt.Sprintf("admin-rest-%s@example.com", suffix), "secret123", "admin")
    client := registerUser(t, "Client Rest", fmt.Sprintf("client-rest-%s@example.com", suffix), "secret123", "client")

    // Admin crea restaurante -> 201
    createResp := doJSONRequest(t, http.MethodPost, "/api/restaurants", map[string]any{
        "name":        "Restaurante " + suffix,
        "address":     "Cartago",
        "phone":       "8888-0000",
        "description": "e2e",
        "capacity":    15,
    }, admin.Token)
    restaurant := decodeJSON[restaurantResponse](t, createResp)
    if createResp.StatusCode != http.StatusCreated {
        t.Fatalf("admin create restaurant debía dar 201 y dio %d", createResp.StatusCode)
    }

    // Usuario normal intenta crear restaurante -> 403
    resp := doJSONRequest(t, http.MethodPost, "/api/restaurants", map[string]any{
        "name":        "No debería",
        "address":     "Cartago",
        "phone":       "8888-9999",
        "description": "e2e",
        "capacity":    10,
    }, client.Token)
    requireStatus(t, resp, http.StatusForbidden)

    // Listar restaurantes sin token -> 200
    resp = doJSONRequest(t, http.MethodGet, "/api/restaurants", nil, "")
    list := decodeJSON[restaurantsListResponse](t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("listar restaurantes público debía dar 200 y dio %d", resp.StatusCode)
    }
    if len(list.Items) == 0 {
        t.Fatal("listar restaurantes devolvió 0 items")
    }

    // Obtener restaurante por ID -> 200
    resp = doJSONRequest(t, http.MethodGet, "/api/restaurants/"+restaurant.ID, nil, "")
    got := decodeJSON[restaurantResponse](t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("get restaurant debía dar 200 y dio %d", resp.StatusCode)
    }
    if got.ID != restaurant.ID {
        t.Fatalf("restaurant incorrecto: %#v", got)
    }

    // ID inexistente -> 404
    resp = doJSONRequest(t, http.MethodGet, "/api/restaurants/no-existe", nil, "")
    requireStatus(t, resp, http.StatusNotFound)
}

func TestE2EMongoMenuAndProductFlows(t *testing.T) {
    suffix := uniqueSuffix()
    admin := registerUser(t, "Admin Menu", fmt.Sprintf("admin-menu-%s@example.com", suffix), "secret123", "admin")
    client := registerUser(t, "Client Menu", fmt.Sprintf("client-menu-%s@example.com", suffix), "secret123", "client")
    restaurant := mustCreateRestaurant(t, admin.Token, suffix)

    // Admin crea menú con productos -> 201
    resp := doJSONRequest(t, http.MethodPost, "/api/menus", map[string]any{
        "restaurant_id": restaurant.ID,
        "name":          "Menu " + suffix,
        "description":   "menu e2e",
        "products": []map[string]any{
            {"name": "Pizza " + suffix, "description": "rica", "category": "pizzas-" + suffix, "price": 4500, "available": true},
            {"name": "Refresco " + suffix, "description": "frío", "category": "bebidas-" + suffix, "price": 1800, "available": true},
        },
    }, admin.Token)
    menu := decodeJSON[menuResponse](t, resp)
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("create menu debía dar 201 y dio %d", resp.StatusCode)
    }
    if len(menu.Products) != 2 {
        t.Fatalf("menú creado debía traer 2 productos y trajo %d", len(menu.Products))
    }

    // Obtener menú por ID -> 200
    resp = doJSONRequest(t, http.MethodGet, "/api/menus/"+menu.ID, nil, admin.Token)
    gotMenu := decodeJSON[menuResponse](t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("get menu debía dar 200 y dio %d", resp.StatusCode)
    }
    if gotMenu.ID != menu.ID {
        t.Fatalf("menú incorrecto: %#v", gotMenu)
    }

    // Admin actualiza menú -> 200
    resp = doJSONRequest(t, http.MethodPut, "/api/menus/"+menu.ID, map[string]any{
        "name":        "Menu actualizado " + suffix,
        "description": "nuevo",
    }, admin.Token)
    updatedMenu := decodeJSON[menuResponse](t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("update menu debía dar 200 y dio %d", resp.StatusCode)
    }
    if updatedMenu.Name != "Menu actualizado "+suffix {
        t.Fatalf("menú no se actualizó: %#v", updatedMenu)
    }

    // Usuario normal intenta actualizar menú -> 403
    resp = doJSONRequest(t, http.MethodPut, "/api/menus/"+menu.ID, map[string]any{
        "name": "No debería",
    }, client.Token)
    requireStatus(t, resp, http.StatusForbidden)

    // Producto por ID -> 200
    prod := menu.Products[0]
    resp = doJSONRequest(t, http.MethodGet, "/api/products/"+prod.ID, nil, admin.Token)
    gotProd := decodeJSON[productResponse](t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("get product debía dar 200 y dio %d", resp.StatusCode)
    }
    if gotProd.ID != prod.ID {
        t.Fatalf("producto incorrecto: %#v", gotProd)
    }

    // Listar productos por categoría -> 200
    resp = doJSONRequest(t, http.MethodGet, "/api/products?category="+prod.Category, nil, admin.Token)
    body := readBody(t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("list by category debía dar 200 y dio %d. body=%s", resp.StatusCode, string(body))
    }
    if !strings.Contains(string(body), prod.ID) {
        t.Fatalf("la categoría no devolvió el producto esperado. body=%s", string(body))
    }

    // Admin actualiza producto -> 200
    resp = doJSONRequest(t, http.MethodPatch, "/api/products/"+prod.ID, map[string]any{
        "price":     4999,
        "available": false,
    }, admin.Token)
    patched := decodeJSON[productResponse](t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("patch product debía dar 200 y dio %d", resp.StatusCode)
    }
    if patched.Price != 4999 || patched.Available {
        t.Fatalf("producto no se actualizó bien: %#v", patched)
    }

    // Usuario normal intenta eliminar producto -> 403
    resp = doJSONRequest(t, http.MethodDelete, "/api/products/"+prod.ID, nil, client.Token)
    requireStatus(t, resp, http.StatusForbidden)

    // Admin elimina producto -> 204
    resp = doJSONRequest(t, http.MethodDelete, "/api/products/"+prod.ID, nil, admin.Token)
    requireStatus(t, resp, http.StatusNoContent)

    // Admin elimina menú -> 204
    resp = doJSONRequest(t, http.MethodDelete, "/api/menus/"+menu.ID, nil, admin.Token)
    requireStatus(t, resp, http.StatusNoContent)

    // Obtener menú eliminado -> 404
    resp = doJSONRequest(t, http.MethodGet, "/api/menus/"+menu.ID, nil, admin.Token)
    requireStatus(t, resp, http.StatusNotFound)
}

func TestE2EMongoReservationAndOrderFlows(t *testing.T) {
    suffix := uniqueSuffix()
    admin := registerUser(t, "Admin Ops", fmt.Sprintf("admin-ops-%s@example.com", suffix), "secret123", "admin")
    user1 := registerUser(t, "User One", fmt.Sprintf("user1-ops-%s@example.com", suffix), "secret123", "client")
    user2 := registerUser(t, "User Two", fmt.Sprintf("user2-ops-%s@example.com", suffix), "secret123", "client")

    restaurant := mustCreateRestaurant(t, admin.Token, suffix)
    menu := mustCreateMenu(t, admin.Token, restaurant.ID, suffix)
    product := menu.Products[0]

    // Usuario crea reservación -> 201
    resp := doJSONRequest(t, http.MethodPost, "/api/reservations", map[string]any{
        "restaurant_id": restaurant.ID,
        "date":          time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
        "party_size":    2,
        "notes":         "mesa junto a ventana",
    }, user1.Token)
    reservation := decodeJSON[reservationResponse](t, resp)
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("create reservation debía dar 201 y dio %d", resp.StatusCode)
    }

    // Usuario intenta cancelar reservación ajena -> 403
    resp = doJSONRequest(t, http.MethodDelete, "/api/reservations/"+reservation.ID, nil, user2.Token)
    requireStatus(t, resp, http.StatusForbidden)

    // Usuario cancela su propia reservación -> 204
    resp = doJSONRequest(t, http.MethodDelete, "/api/reservations/"+reservation.ID, nil, user1.Token)
    requireStatus(t, resp, http.StatusNoContent)

    // Con el código actual esto devuelve 422, no 404, porque el service lo trata como ErrValidation.
    resp = doJSONRequest(t, http.MethodPost, "/api/reservations", map[string]any{
        "restaurant_id": "no-existe",
        "date":          time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339),
        "party_size":    2,
    }, user1.Token)
    requireStatus(t, resp, http.StatusUnprocessableEntity)

    // Usuario crea orden con productos válidos -> 201
    resp = doJSONRequest(t, http.MethodPost, "/api/orders", map[string]any{
        "restaurant_id": restaurant.ID,
        "items": []map[string]any{{
            "product_id": product.ID,
            "quantity":   2,
        }},
        "pickup": true,
    }, user1.Token)
    order := decodeJSON[orderResponse](t, resp)
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("create order debía dar 201 y dio %d", resp.StatusCode)
    }
    if len(order.Items) != 1 || order.Total <= 0 {
        t.Fatalf("orden incorrecta: %#v", order)
    }

    // Obtener orden por ID -> 200
    resp = doJSONRequest(t, http.MethodGet, "/api/orders/"+order.ID, nil, user1.Token)
    gotOrder := decodeJSON[orderResponse](t, resp)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("get order debía dar 200 y dio %d", resp.StatusCode)
    }
    if gotOrder.ID != order.ID {
        t.Fatalf("orden incorrecta: %#v", gotOrder)
    }

    // Con el código actual esto devuelve 422, no 404, porque el service lo trata como ErrValidation.
    resp = doJSONRequest(t, http.MethodPost, "/api/orders", map[string]any{
        "restaurant_id": restaurant.ID,
        "items": []map[string]any{{
            "product_id": "prod-no-existe",
            "quantity":   1,
        }},
    }, user1.Token)
    requireStatus(t, resp, http.StatusUnprocessableEntity)
}
