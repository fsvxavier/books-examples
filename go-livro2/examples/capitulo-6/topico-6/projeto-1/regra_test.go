package main

import (
	"errors"
	"strings"
	"testing"
)

func TestTodasReuneErros(t *testing.T) {
	r := Todas(NaoVazio("x"), Tamanho("x", 3, 5), Contem("x", "@"))
	err := r.Validar("")
	if err == nil {
		t.Fatal("esperava erros")
	}
	for _, quero := range []string{"x: obrigatorio", "tamanho 0", `deve conter "@"`} {
		if !strings.Contains(err.Error(), quero) {
			t.Errorf("faltou %q em:\n%v", quero, err)
		}
	}
}

func TestAlgumaPassaComUma(t *testing.T) {
	r := Alguma(Tamanho("a", 1, 1), Tamanho("b", 3, 3))
	if err := r.Validar("abc"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if err := r.Validar("ab"); err == nil {
		t.Fatal("esperava falha")
	}
}

func TestNao(t *testing.T) {
	r := Nao(Contem("x", "@"), "nao pode ter @")
	if r.Validar("abc") != nil || r.Validar("a@b") == nil {
		t.Fatal("Nao nao inverteu")
	}
}

// A composicao tambem e uma Regra: da para aninhar sem limite.
func TestAninhamento(t *testing.T) {
	r := Todas(NaoVazio("x"), Alguma(Contem("x", "a"), Todas(Contem("x", "b"), Contem("x", "c"))))
	cases := map[string]bool{"a": true, "bc": true, "b": false, "": false, "abc": true}
	for in, ok := range cases {
		if got := r.Validar(in) == nil; got != ok {
			t.Errorf("Validar(%q) passou=%v, quero %v", in, got, ok)
		}
	}
}

func TestErroOriginalContinuaNaCadeia(t *testing.T) {
	sentinela := errors.New("proibido")
	r := Todas(RegraFunc[int](func(int) error { return sentinela }))
	if !errors.Is(r.Validar(1), sentinela) {
		t.Fatal("errors.Is nao achou o erro dentro da composicao")
	}
}

func TestVazioPassa(t *testing.T) {
	if err := Todas[string]().Validar("x"); err != nil {
		t.Fatalf("Todas sem regras deveria passar: %v", err)
	}
}
