package main

import (
	"errors"
	"math/rand/v2"
	"testing"
)

func TestDesfazerERefazer(t *testing.T) {
	ed := NovoEditor("abc")
	h := &Historico{}
	h.Executar(&InserirTexto{ed, 3, "def"})
	h.Executar(&ApagarTrecho{Ed: ed, Pos: 0, N: 2})
	if ed.Texto() != "cdef" {
		t.Fatalf("texto = %q", ed.Texto())
	}
	h.Desfazer()
	h.Desfazer()
	if ed.Texto() != "abc" {
		t.Fatalf("depois de desfazer: %q", ed.Texto())
	}
	h.Refazer()
	h.Refazer()
	if ed.Texto() != "cdef" {
		t.Fatalf("depois de refazer: %q", ed.Texto())
	}
}

func TestErrosDeHistoricoVazio(t *testing.T) {
	h := &Historico{}
	if !errors.Is(h.Desfazer(), ErroNadaParaDesfazer) || !errors.Is(h.Refazer(), ErroNadaParaRefazer) {
		t.Fatal("erros de historico vazio errados")
	}
}

func TestComandoQueFalhaNaoEntraNoHistorico(t *testing.T) {
	ed := NovoEditor("abc")
	h := &Historico{}
	if err := h.Executar(&ApagarTrecho{Ed: ed, Pos: 10, N: 1}); err == nil {
		t.Fatal("esperava erro")
	}
	if f, _ := h.Tamanho(); f != 0 || ed.Texto() != "abc" {
		t.Fatalf("feitos=%d texto=%q", f, ed.Texto())
	}
}

func TestComandoNovoLimpaORefazer(t *testing.T) {
	ed := NovoEditor("")
	h := &Historico{}
	h.Executar(&InserirTexto{ed, 0, "a"})
	h.Desfazer()
	h.Executar(&InserirTexto{ed, 0, "b"})
	if err := h.Refazer(); !errors.Is(err, ErroNadaParaRefazer) {
		t.Fatalf("err = %v", err)
	}
}

func TestLimiteDoHistorico(t *testing.T) {
	ed := NovoEditor("")
	h := &Historico{Limite: 3}
	for range 10 {
		h.Executar(&InserirTexto{ed, 0, "x"})
	}
	if f, _ := h.Tamanho(); f != 3 {
		t.Fatalf("feitos = %d", f)
	}
	for h.Desfazer() == nil {
	}
	if len([]rune(ed.Texto())) != 7 {
		t.Fatalf("texto = %q: so os 3 ultimos puderam ser desfeitos", ed.Texto())
	}
}

func TestUnicode(t *testing.T) {
	ed := NovoEditor("coração")
	h := &Historico{}
	h.Executar(&InserirTexto{ed, 4, "日本"}) // posicao em CARACTERES, nao em bytes
	h.Executar(&Maiusculas{Ed: ed})
	h.Desfazer()
	h.Desfazer()
	if ed.Texto() != "coração" {
		t.Fatalf("texto = %q", ed.Texto())
	}
}

// Propriedade: para QUALQUER sequencia valida de comandos, desfazer tudo volta ao inicio e
// refazer tudo volta ao fim. Sequencias aleatorias com semente fixa (reproduzivel).
func TestPropriedadeIdaEVolta(t *testing.T) {
	r := rand.New(rand.NewPCG(42, 99))
	for iter := range 500 {
		inicial := "ação é importante"
		ed := NovoEditor(inicial)
		h := &Historico{}
		for range r.IntN(30) {
			n := len(ed.texto)
			var c Comando
			switch r.IntN(3) {
			case 0:
				c = &InserirTexto{ed, r.IntN(n + 1), []string{"x", "é", "日本", "AbC"}[r.IntN(4)]}
			case 1:
				if n == 0 {
					continue
				}
				pos := r.IntN(n)
				c = &ApagarTrecho{Ed: ed, Pos: pos, N: r.IntN(n - pos + 1)}
			default:
				c = &Maiusculas{Ed: ed}
			}
			if err := h.Executar(c); err != nil {
				t.Fatalf("iteracao %d: %v", iter, err)
			}
		}
		final := ed.Texto()
		for h.Desfazer() == nil {
		}
		if ed.Texto() != inicial {
			t.Fatalf("iteracao %d: desfazer tudo deu %q, quero %q", iter, ed.Texto(), inicial)
		}
		for h.Refazer() == nil {
		}
		if ed.Texto() != final {
			t.Fatalf("iteracao %d: refazer tudo deu %q, quero %q", iter, ed.Texto(), final)
		}
	}
}
