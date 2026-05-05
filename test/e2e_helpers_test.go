//go:build e2e

package test

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "testing"
    "time"
)

const baseURL = "http://localhost"

var e2eClient = &http.Client{Timeout: 20 * time.Second}

type e2eUser struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Role  string `json:"role"`
}

type authResponse struct {
    Token string  `json:"token"`
    User  e2eUser `json:"user"`
}

type restaurantResponse struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    AdminID  string `json:"admin_id"`
    Capacity int    `json:"capacity"`
}

type restaurantsListResponse struct {
    Items []restaurantResponse `json:"items"`
}

type menuResponse struct {
    ID           string            `json:"id"`
    RestaurantID string            `json:"restaurant_id"`
    Name         string            `json:"name"`
    Products     []productResponse `json:"products"`
}

type productResponse struct {
    ID           string  `json:"id"`
    MenuID       string  `json:"menu_id"`
    RestaurantID string  `json:"restaurant_id"`
    Name         string  `json:"name"`
    Description  string  `json:"description"`
    Category     string  `json:"category"`
    Price        float64 `json:"price"`
    Available    bool    `json:"available"`
}

type reservationResponse struct {
    ID           string `json:"id"`
    RestaurantID string `json:"restaurant_id"`
    UserID       string `json:"user_id"`
    Status       string `json:"status"`
}

type orderItemResponse struct {
    ProductID string  `json:"product_id"`
    Quantity  int     `json:"quantity"`
    Price     float64 `json:"price"`
}

type orderResponse struct {
    ID           string              `json:"id"`
    UserID       string              `json:"user_id"`
    RestaurantID string              `json:"restaurant_id"`
    Items        []orderItemResponse `json:"items"`
    Total        float64             `json:"total"`
}

type searchResponse struct {
    Query    string            `json:"query"`
    Category string            `json:"category"`
    Count    int               `json:"count"`
    Items    []productResponse `json:"items"`
}

func uniqueSuffix() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}

func doJSONRequest(t *testing.T, method, path string, body any, token string) *http.Response {
    t.Helper()

    var reader io.Reader
    if body != nil {
        raw, err := json.Marshal(body)
        if err != nil {
            t.Fatalf("no pude serializar body: %v", err)
        }
        reader = bytes.NewReader(raw)
    }

    req, err := http.NewRequest(method, baseURL+path, reader)
    if err != nil {
        t.Fatalf("no pude crear request: %v", err)
    }
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }

    resp, err := e2eClient.Do(req)
    if err != nil {
        t.Fatalf("request falló: %v", err)
    }
    return resp
}

func readBody(t *testing.T, resp *http.Response) []byte {
    t.Helper()
    defer resp.Body.Close()
    raw, err := io.ReadAll(resp.Body)
    if err != nil {
        t.Fatalf("no pude leer body: %v", err)
    }
    return raw
}

func decodeJSON[T any](t *testing.T, resp *http.Response) T {
    t.Helper()
    raw := readBody(t, resp)
    var out T
    if err := json.Unmarshal(raw, &out); err != nil {
        t.Fatalf("json inválido: %v body=%s", err, string(raw))
    }
    return out
}

func requireStatus(t *testing.T, resp *http.Response, expected int) {
    t.Helper()
    raw := readBody(t, resp)
    if resp.StatusCode != expected {
        t.Fatalf("status esperado %d, obtenido %d. body=%s", expected, resp.StatusCode, string(raw))
    }
}

func registerUser(t *testing.T, name, email, password, role string) authResponse {
    t.Helper()
    resp := doJSONRequest(t, http.MethodPost, "/auth/register", map[string]any{
        "name":     name,
        "email":    email,
        "password": password,
        "role":     role,
    }, "")
    if resp.StatusCode != http.StatusCreated {
        raw := readBody(t, resp)
        t.Fatalf("registro falló: status=%d body=%s", resp.StatusCode, string(raw))
    }
    return decodeJSON[authResponse](t, resp)
}

func loginUser(t *testing.T, email, password string) authResponse {
    t.Helper()
    resp := doJSONRequest(t, http.MethodPost, "/auth/login", map[string]any{
        "email":    email,
        "password": password,
    }, "")
    if resp.StatusCode != http.StatusOK {
        raw := readBody(t, resp)
        t.Fatalf("login falló: status=%d body=%s", resp.StatusCode, string(raw))
    }
    return decodeJSON[authResponse](t, resp)
}

func mustCreateRestaurant(t *testing.T, token, suffix string) restaurantResponse {
    t.Helper()
    resp := doJSONRequest(t, http.MethodPost, "/api/restaurants", map[string]any{
        "name":        "Rest " + suffix,
        "address":     "Cartago",
        "phone":       "5555-" + suffix[len(suffix)-4:],
        "description": "E2E",
        "capacity":    20,
    }, token)
    if resp.StatusCode != http.StatusCreated {
        raw := readBody(t, resp)
        t.Fatalf("create restaurant falló: status=%d body=%s", resp.StatusCode, string(raw))
    }
    return decodeJSON[restaurantResponse](t, resp)
}

func mustCreateMenu(t *testing.T, token, restaurantID, suffix string) menuResponse {
    t.Helper()
    resp := doJSONRequest(t, http.MethodPost, "/api/menus", map[string]any{
        "restaurant_id": restaurantID,
        "name":          "Menu " + suffix,
        "description":   "menu e2e",
        "products": []map[string]any{
            {
                "name":        "Pizza " + suffix,
                "description": "Pizza de prueba",
                "category":    "cat-" + suffix,
                "price":       4500,
                "available":   true,
            },
            {
                "name":        "Bebida " + suffix,
                "description": "",
                "category":    "bebidas-" + suffix,
                "price":       1800,
                "available":   true,
            },
        },
    }, token)
    if resp.StatusCode != http.StatusCreated {
        raw := readBody(t, resp)
        t.Fatalf("create menu falló: status=%d body=%s", resp.StatusCode, string(raw))
    }
    return decodeJSON[menuResponse](t, resp)
}

func containsProductByName(items []productResponse, name string) bool {
    for _, item := range items {
        if strings.EqualFold(item.Name, name) {
            return true
        }
    }
    return false
}
