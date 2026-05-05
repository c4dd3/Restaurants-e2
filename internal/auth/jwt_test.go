package auth

import (
	"testing"
	"time"

	"restaurants-e2/internal/domain"
)

func TestSignAndParse(t *testing.T) {
	secret := "secret-de-prueba-suficientemente-largo"
	token, err := Sign("user-42", "test@example.com", domain.RoleAdmin, secret, time.Hour)
	if err != nil {
		t.Fatalf("Sign inesperado: %v", err)
	}
	if token == "" {
		t.Fatal("token vacío")
	}

	claims, err := Parse(token, secret)
	if err != nil {
		t.Fatalf("Parse inesperado: %v", err)
	}
	if claims.UserID != "user-42" {
		t.Errorf("UserID esperado user-42, obtenido %q", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Email esperado test@example.com, obtenido %q", claims.Email)
	}
	if claims.Role != domain.RoleAdmin {
		t.Errorf("Role esperado %q, obtenido %q", domain.RoleAdmin, claims.Role)
	}
	if claims.Issuer != "restaurants-e2" {
		t.Errorf("Issuer esperado restaurants-e2, obtenido %q", claims.Issuer)
	}
}

func TestParseExpiredToken(t *testing.T) {
	secret := "secret-de-prueba"
	token, _ := Sign("user-1", "x@x.com", domain.RoleClient, secret, -time.Second)

	_, err := Parse(token, secret)
	if err == nil {
		t.Fatal("Parse aceptó un token expirado")
	}
}

func TestParseWrongSecret(t *testing.T) {
	token, _ := Sign("user-1", "x@x.com", domain.RoleClient, "secret-original", time.Hour)

	_, err := Parse(token, "secret-diferente")
	if err == nil {
		t.Fatal("Parse aceptó token firmado con secret diferente")
	}
}

func TestParseInvalidToken(t *testing.T) {
	_, err := Parse("esto.no.es.un.jwt", "cualquier-secret")
	if err == nil {
		t.Fatal("Parse aceptó string inválido como JWT")
	}
}

func TestParseAlgNoneAttack(t *testing.T) {
	malicious := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0" +
		".eyJzdWIiOiJhdHRhY2tlciIsImVtYWlsIjoiYXR0YWNrZXJAZXhhbXBsZS5jb20iLCJyb2xlIjoiYWRtaW4ifQ" +
		"."

	_, err := Parse(malicious, "cualquier-secret")
	if err == nil {
		t.Fatal("Parse aceptó token con alg=none (ataque alg-substitution)")
	}
}
