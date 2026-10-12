package main

import (
	"strings"
	"testing"
)

func TestCodigoAleatorio(t *testing.T) {
	vistos := make(map[string]bool)
	for range 10000 {
		c := codigoAleatorio()
		if len(c) != tamanhoCodigo {
			t.Fatalf("tamanho %d", len(c))
		}
		for _, r := range c {
			if !strings.ContainsRune(alfabeto, r) {
				t.Fatalf("símbolo %q fora do alfabeto", r)
			}
		}
		vistos[c] = true
	}
	// 32^6 = 1.073.741.824 códigos: 10.000 sorteios quase nunca repetem.
	if len(vistos) < 9990 {
		t.Errorf("só %d códigos distintos em 10000", len(vistos))
	}
}

func TestValidarDestino(t *testing.T) {
	bons := map[string]string{
		"https://go.dev/doc/":     "https://go.dev/doc/",
		"  http://exemplo.com/a ": "http://exemplo.com/a",
	}
	for entrada, esperado := range bons {
		got, err := validarDestino(entrada)
		if err != nil || got != esperado {
			t.Errorf("validarDestino(%q) = %q, %v", entrada, got, err)
		}
	}
	for _, ruim := range []string{"", "go.dev", "javascript:alert(1)", "http://", "http://x.org/" + strings.Repeat("a", 3000)} {
		if _, err := validarDestino(ruim); err == nil {
			t.Errorf("validarDestino(%.30q) deveria falhar", ruim)
		}
	}
}
