package main

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

var itensTeste = []Item{{"a", 10000, 500}, {"b", 5000, 500}}

func novo(t *testing.T) *Pedido {
	t.Helper()
	p, err := NovoPedido("T", FreteFixo, itensTeste...)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSemItens(t *testing.T) {
	if _, err := NovoPedido("x", FreteFixo); !errors.Is(err, ErrSemItens) {
		t.Fatalf("err = %v", err)
	}
}

func TestMaquinaCompleta(t *testing.T) {
	if l := novo(t).maq.Lacunas(todosEstados, todosEventos); len(l) != 0 {
		t.Fatalf("lacunas: %v", l)
	}
}

// Tabela com todos os caminhos permitidos e uma amostra dos proibidos.
func TestTransicoes(t *testing.T) {
	casos := []struct {
		nome    string
		eventos []Evento
		final   Est
		falha   int // índice do primeiro evento recusado, -1 = nenhum
	}{
		{"feliz", []Evento{EvPagar, EvEnviar, EvEntregar}, Entregue, -1},
		{"cancela criado", []Evento{EvCancelar}, Cancelado, -1},
		{"cancela pago", []Evento{EvPagar, EvCancelar}, Cancelado, -1},
		{"cancela enviado", []Evento{EvPagar, EvEnviar, EvCancelar}, Enviado, 2},
		{"enviar sem pagar", []Evento{EvEnviar}, Criado, 0},
		{"pagar duas vezes", []Evento{EvPagar, EvPagar}, Pago, 1},
		{"depois de cancelado", []Evento{EvCancelar, EvPagar}, Cancelado, 1},
		{"depois de entregue", []Evento{EvPagar, EvEnviar, EvEntregar, EvCancelar}, Entregue, 3},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			p := novo(t)
			falhou := -1
			for i, ev := range c.eventos {
				if err := p.maq.Enviar(ev); err != nil {
					if !errors.Is(err, ErrTransicao) {
						t.Fatalf("erro inesperado: %v", err)
					}
					if falhou < 0 {
						falhou = i
					}
				}
			}
			if falhou != c.falha || p.Estado() != c.final {
				t.Fatalf("falhou=%d estado=%s; quero %d/%s", falhou, p.Estado(), c.falha, c.final)
			}
		})
	}
}

func TestReembolsoDependeDoEstado(t *testing.T) {
	a := novo(t)
	_ = a.Cancelar()
	b := novo(t)
	_ = b.Pagar()
	_ = b.Cancelar()
	if a.Reembolso() != 0 {
		t.Fatalf("criado: %d", a.Reembolso())
	}
	if want := b.Total() - b.Total()/10; b.Reembolso() != want {
		t.Fatalf("pago: %d, quero %d", b.Reembolso(), want)
	}
}

func TestFretes(t *testing.T) {
	if FreteFixo(0, 0) != 2000 {
		t.Fatal("fixo")
	}
	if FretePorKg(1500, 0) != 1800 { // 800 + 2 kg * 500
		t.Fatal("por kg")
	}
	g := FreteGratis(15000)
	if g.Calcular(1500, 15000) != 0 || g.Calcular(1500, 14999) != 1800 {
		t.Fatal("grátis")
	}
}

func TestEstadoTexto(t *testing.T) {
	var e Est
	if err := e.UnmarshalText([]byte("enviado")); err != nil || e != Enviado {
		t.Fatalf("%v %v", e, err)
	}
	if err := e.UnmarshalText([]byte("x")); err == nil {
		t.Fatal("esperava erro")
	}
}

// Pagar e Cancelar ao mesmo tempo: exatamente um estado final coerente e
// um reembolso coerente com ele, em qualquer ordem de chegada.
func TestPagarECancelarConcorrentes(t *testing.T) {
	var cancelouAntes, cancelouDepois atomic.Int32
	for range 500 {
		p := novo(t)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _ = p.Pagar() }()
		go func() { defer wg.Done(); _ = p.Cancelar() }()
		wg.Wait()
		if p.Estado() != Cancelado {
			t.Fatalf("estado %s: cancelar é permitido em criado e em pago, então sempre vence", p.Estado())
		}
		// Se o cancelamento chegou antes do pagamento, o reembolso é 0 (e
		// o pagamento foi recusado); se chegou depois, o reembolso retém a taxa.
		if p.Reembolso() == 0 {
			cancelouAntes.Add(1)
		} else {
			cancelouDepois.Add(1)
		}
	}
	t.Logf("cancelou antes de pagar (reembolso 0): %d; depois de pagar (reembolso com taxa): %d",
		cancelouAntes.Load(), cancelouDepois.Load())
}
