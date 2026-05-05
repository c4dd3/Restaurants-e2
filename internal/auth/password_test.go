package auth

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	plain := "supersecret123"

	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword inesperado: %v", err)
	}
	if hash == "" || hash == plain {
		t.Fatal("hash vacío o igual al texto claro")
	}

	if err := VerifyPassword(hash, plain); err != nil {
		t.Fatalf("VerifyPassword rechazó contraseña válida: %v", err)
	}
}

func TestVerifyPasswordWrong(t *testing.T) {
	hash, _ := HashPassword("correct-password")

	if err := VerifyPassword(hash, "wrong-password"); err == nil {
		t.Fatal("VerifyPassword aceptó contraseña incorrecta")
	}
}

func TestHashesAreDifferentEachCall(t *testing.T) {
	// bcrypt genera salt aleatorio; dos hashes del mismo texto deben ser distintos.
	h1, _ := HashPassword("misma-pass")
	h2, _ := HashPassword("misma-pass")
	if h1 == h2 {
		t.Fatal("dos hashes del mismo texto son idénticos — bcrypt no está generando salt")
	}
}
