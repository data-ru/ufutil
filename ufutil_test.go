package ufutil

import "testing"

func TestLogin(t *testing.T) {
	loginer := Usuario{
		Usuario: "",
		Senha:   "",
	}
	a, err := Login(loginer)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a.Cpf, a.Nome)
}

func TestDecifrarComoTexto(t *testing.T) {
	a, err := decryptJsonAsString("")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	t.Log(a)
}

func TestDecifrarComoStruct(t *testing.T) {
	a, err := decryptJSON("")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	t.Log(a)
}
