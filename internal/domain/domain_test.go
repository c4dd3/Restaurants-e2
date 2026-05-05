package domain

import "testing"

func TestProductDescriptionOrDefault(t *testing.T) {
	p := &Product{Name: "Pizza"}
	if p.DescriptionOrDefault() != DefaultProductDescription {
		t.Fatalf("esperaba descripción por defecto, obtuvo %q", p.DescriptionOrDefault())
	}

	p.Description = "Deliciosa pizza"
	if p.DescriptionOrDefault() != "Deliciosa pizza" {
		t.Fatalf("esperaba descripción custom, obtuvo %q", p.DescriptionOrDefault())
	}
}

func TestUserIsAdmin(t *testing.T) {
	admin := &User{Role: RoleAdmin}
	if !admin.IsAdmin() {
		t.Fatal("usuario con RoleAdmin debería ser admin")
	}

	client := &User{Role: RoleClient}
	if client.IsAdmin() {
		t.Fatal("usuario con RoleClient no debería ser admin")
	}
}
