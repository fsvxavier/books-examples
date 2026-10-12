package main

import (
	"math/rand/v2"
	"testing"
)

// Propriedade: depois de qualquer sequência de edições, desfazer tudo
// (dentro do limite) devolve exatamente os estados anteriores, e refazer
// tudo devolve o estado final.
func TestDesfazerRefazerPropriedade(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 11))
	for rodada := range 500 {
		ed := &Editor{}
		h := NovoHistorico(1000)
		var estados []Estado
		n := 1 + r.IntN(30)
		for range n {
			estados = append(estados, ed.Salvar())
			h.Registrar(ed.Salvar())
			if r.IntN(3) == 0 {
				ed.Apagar(1 + r.IntN(3))
			} else {
				ed.Inserir(string(rune('a' + r.IntN(26))))
			}
		}
		final := ed.Salvar()

		for i := n - 1; i >= 0; i-- {
			s, ok := h.Desfazer(ed.Salvar())
			if !ok {
				t.Fatalf("rodada %d: desfazer %d falhou", rodada, i)
			}
			ed.Restaurar(s)
			if ed.Salvar() != estados[i] {
				t.Fatalf("rodada %d: passo %d: %v != %v", rodada, i, ed.Salvar(), estados[i])
			}
		}
		for range n {
			s, ok := h.Refazer(ed.Salvar())
			if !ok {
				t.Fatalf("rodada %d: refazer falhou", rodada)
			}
			ed.Restaurar(s)
		}
		if ed.Salvar() != final {
			t.Fatalf("rodada %d: final %v != %v", rodada, ed.Salvar(), final)
		}
	}
}

func TestLimiteDescartaOMaisAntigo(t *testing.T) {
	ed := &Editor{}
	h := NovoHistorico(2)
	for _, s := range []string{"a", "b", "c"} {
		h.Registrar(ed.Salvar())
		ed.Inserir(s)
	}
	desfeitos := 0
	for {
		s, ok := h.Desfazer(ed.Salvar())
		if !ok {
			break
		}
		ed.Restaurar(s)
		desfeitos++
	}
	if desfeitos != 2 || ed.Salvar().Texto != "a" {
		t.Fatalf("desfeitos=%d texto=%q", desfeitos, ed.Salvar().Texto)
	}
}
