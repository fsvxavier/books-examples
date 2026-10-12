package main

import (
	"errors"
	"testing"
)

// Teste de exaustividade: todo tipo de nó precisa ser tratado por Descrever.
func TestDescreverTrataTodosOsTipos(t *testing.T) {
	for _, n := range Amostras() {
		if _, err := Descrever(n); errors.Is(err, ErrNaoTratado) {
			t.Errorf("Descrever não trata %T: acrescente um case", n)
		}
	}
}
