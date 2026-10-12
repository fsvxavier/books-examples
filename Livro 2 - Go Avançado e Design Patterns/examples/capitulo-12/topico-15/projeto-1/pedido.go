package main

import (
	"errors"
	"fmt"
	"sync"
)

// Est é o estado do pedido. Implementa encoding.TextMarshaler para
// aparecer como texto no JSON e no banco.
type Est int

const (
	Criado Est = iota
	Pago
	Enviado
	Entregue
	Cancelado
)

var nomesEst = [...]string{"criado", "pago", "enviado", "entregue", "cancelado"}

func (e Est) String() string { return nomesEst[e] }

func (e Est) MarshalText() ([]byte, error) { return []byte(e.String()), nil }

func (e *Est) UnmarshalText(b []byte) error {
	for i, n := range nomesEst {
		if n == string(b) {
			*e = Est(i)
			return nil
		}
	}
	return fmt.Errorf("estado desconhecido %q", b)
}

type Evento int

const (
	EvPagar Evento = iota
	EvEnviar
	EvEntregar
	EvCancelar
)

var nomesEv = [...]string{"pagar", "enviar", "entregar", "cancelar"}

func (e Evento) String() string { return nomesEv[e] }

var todosEstados = []Est{Criado, Pago, Enviado, Entregue, Cancelado}
var todosEventos = []Evento{EvPagar, EvEnviar, EvEntregar, EvCancelar}

var proibidos = map[Est][]Evento{
	Criado:    {EvEnviar, EvEntregar},
	Pago:      {EvPagar, EvEntregar},
	Enviado:   {EvPagar, EvEnviar, EvCancelar}, // cancelar enviado exige devolução
	Entregue:  {EvPagar, EvEnviar, EvEntregar, EvCancelar},
	Cancelado: {EvPagar, EvEnviar, EvEntregar, EvCancelar},
}

// Frete é uma estratégia: escolhida por quem cria o pedido.
type Frete interface {
	Calcular(pesoGramas int, subtotal int) int
}

type FreteFunc func(pesoGramas, subtotal int) int

func (f FreteFunc) Calcular(p, s int) int { return f(p, s) }

var (
	FreteFixo   = FreteFunc(func(int, int) int { return 2000 })
	FretePorKg  = FreteFunc(func(p, _ int) int { return 800 + (p+999)/1000*500 })
	FreteGratis = func(acima int) Frete {
		return FreteFunc(func(p, s int) int {
			if s >= acima {
				return 0
			}
			return FretePorKg(p, s)
		})
	}
)

type Item struct {
	Nome  string
	Preco int // centavos
	Peso  int // gramas
}

type Pedido struct {
	ID    string
	itens []Item
	frete Frete

	mu        sync.Mutex
	reembolso int

	maq *Maquina[Est, Evento]
}

var ErrSemItens = errors.New("pedido sem itens")

func NovoPedido(id string, frete Frete, itens ...Item) (*Pedido, error) {
	return reidratar(id, Criado, frete, itens)
}

// reidratar monta o pedido em um estado qualquer (vindo do banco).
func reidratar(id string, est Est, frete Frete, itens []Item) (*Pedido, error) {
	if len(itens) == 0 {
		return nil, ErrSemItens
	}
	p := &Pedido{ID: id, itens: itens, frete: frete}
	p.maq = Nova[Est, Evento](est).
		Permitir(Criado, EvPagar, Regra[Est]{Para: Pago}).
		Permitir(Pago, EvEnviar, Regra[Est]{Para: Enviado}).
		Permitir(Enviado, EvEntregar, Regra[Est]{Para: Entregue}).
		// O ESTADO escolhe a política de cancelamento: cada regra tem a sua.
		Permitir(Criado, EvCancelar, Regra[Est]{Para: Cancelado, Acao: func() error {
			p.definirReembolso(0) // nada foi cobrado
			return nil
		}}).
		Permitir(Pago, EvCancelar, Regra[Est]{Para: Cancelado, Acao: func() error {
			total := p.Total()
			p.definirReembolso(total - total/10) // retém 10% de taxa
			return nil
		}})
	// As 15 proibições são decisões, uma a uma: se alguém acrescentar um
	// estado ou um evento e esquecer de decidir, Lacunas acusa.
	for est, eventos := range proibidos {
		for _, ev := range eventos {
			p.maq.Proibir(est, ev)
		}
	}
	return p, nil
}

func (p *Pedido) definirReembolso(v int) {
	p.mu.Lock()
	p.reembolso = v
	p.mu.Unlock()
}

func (p *Pedido) Reembolso() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.reembolso
}

func (p *Pedido) Subtotal() int {
	t := 0
	for _, i := range p.itens {
		t += i.Preco
	}
	return t
}

func (p *Pedido) Peso() int {
	t := 0
	for _, i := range p.itens {
		t += i.Peso
	}
	return t
}

func (p *Pedido) Frete() int { return p.frete.Calcular(p.Peso(), p.Subtotal()) }
func (p *Pedido) Total() int { return p.Subtotal() + p.Frete() }

func (p *Pedido) Estado() Est { return p.maq.Estado() }

func (p *Pedido) Pagar() error    { return p.maq.Enviar(EvPagar) }
func (p *Pedido) Enviar() error   { return p.maq.Enviar(EvEnviar) }
func (p *Pedido) Entregar() error { return p.maq.Enviar(EvEntregar) }
func (p *Pedido) Cancelar() error { return p.maq.Enviar(EvCancelar) }

func (p *Pedido) Observar(f func(Mudanca[Est, Evento])) { p.maq.Observar(f) }
