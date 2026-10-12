package lsp

import (
	"errors"
	"testing"
)

// contratoRetangulo e o que qualquer chamador de Retangulo espera.
func contratoRetangulo(t *testing.T, nome string, r Retangulo) {
	t.Helper()
	r.DefinirLargura(5)
	r.DefinirAltura(4)
	if got := r.Area(); got != 20 {
		t.Errorf("%s: largura 5 x altura 4 deu area %d, esperava 20", nome, got)
	}
}

func TestRetangulo(t *testing.T) {
	contratoRetangulo(t, "Ret", &Ret{})
	contratoRetangulo(t, "Quadrado", &Quadrado{}) // falha de proposito
}

func contratoArmazem(t *testing.T, nome string, a Armazem) {
	t.Helper()
	err := a.Guardar("k", "v")
	if err != nil {
		return // recusar com erro e permitido pelo contrato
	}
	if v, ok := a.Ler("k"); !ok || v != "v" {
		t.Errorf("%s: Guardar devolveu nil, mas Ler nao acha o valor (%q, %v)", nome, v, ok)
	}
}

func TestArmazem(t *testing.T) {
	contratoArmazem(t, "Memoria", NovaMemoria())
	contratoArmazem(t, "SomenteLeituraMentirosa", &SomenteLeituraMentirosa{*NovaMemoria()}) // falha
	sl := SomenteLeitura{NovaMemoria()}
	contratoArmazem(t, "SomenteLeitura", sl)
	if err := sl.Guardar("a", "b"); !errors.Is(err, ErrSomenteLeitura) {
		t.Errorf("esperava ErrSomenteLeitura, veio %v", err)
	}
}

var _ Leitor = SomenteLeitura{} // verificacao em tempo de compilacao
