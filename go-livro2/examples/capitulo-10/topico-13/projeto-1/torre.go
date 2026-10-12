package main

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
)

// Torre e o MEDIADOR: uma goroutine que e DONA do estado da pista. As aeronaves (outras goroutines)
// nao conversam entre si e nao compartilham memoria com ela: so trocam mensagens por canais.
type Torre struct {
	pedidos    chan pedido
	liberacoes chan string
	consultas  chan chan int
	parar      chan struct{}
	feito      chan struct{}
}

type pedido struct {
	id         string
	prioridade int // maior = mais urgente
	ordem      int // desempate: quem chegou primeiro
	resposta   chan struct{}
}

var ErrTorreEncerrada = errors.New("torre encerrada")

func NovaTorre() *Torre {
	t := &Torre{
		pedidos:    make(chan pedido),
		liberacoes: make(chan string),
		consultas:  make(chan chan int),
		parar:      make(chan struct{}),
		feito:      make(chan struct{}),
	}
	go t.executar()
	return t
}

// executar e o UNICO lugar que le e escreve o estado: nao ha Mutex porque nao ha acesso compartilhado.
func (t *Torre) executar() {
	defer close(t.feito)
	var (
		ocupada string // quem esta usando a pista ("" = livre)
		fila    filaPrioridade
		chegada int
	)
	conceder := func() {
		if ocupada == "" && fila.Len() > 0 {
			p := heap.Pop(&fila).(pedido)
			ocupada = p.id
			close(p.resposta) // acorda a aeronave: pode pousar
		}
	}
	for {
		select {
		case p := <-t.pedidos:
			chegada++
			p.ordem = chegada
			heap.Push(&fila, p)
			conceder()
		case id := <-t.liberacoes:
			if ocupada == id {
				ocupada = ""
			}
			conceder()
		case resp := <-t.consultas:
			resp <- fila.Len()
		case <-t.parar:
			return
		}
	}
}

// Solicitar bloqueia ate a pista ser concedida (ou o contexto acabar). O que a aeronave sabe da
// torre e SO este metodo e Liberar: nao sabe quantas outras aeronaves existem.
func (t *Torre) Solicitar(ctx context.Context, id string, prioridade int) error {
	resposta := make(chan struct{})
	select {
	case t.pedidos <- pedido{id: id, prioridade: prioridade, resposta: resposta}:
	case <-ctx.Done():
		return ctx.Err()
	case <-t.feito:
		return ErrTorreEncerrada
	}
	select {
	case <-resposta:
		return nil
	case <-ctx.Done():
		// OBS: o pedido continua na fila da torre; um sistema real teria uma mensagem de cancelamento.
		return fmt.Errorf("%s desistiu: %w", id, ctx.Err())
	case <-t.feito:
		return ErrTorreEncerrada
	}
}

func (t *Torre) Liberar(id string) {
	select {
	case t.liberacoes <- id:
	case <-t.feito:
	}
}

// Esperando consulta quantas aeronaves aguardam (uma mensagem a torre, nao uma leitura da memoria dela).
func (t *Torre) Esperando() int {
	r := make(chan int, 1)
	select {
	case t.consultas <- r:
		return <-r
	case <-t.feito:
		return 0
	}
}

func (t *Torre) Encerrar() { close(t.parar); <-t.feito }

// ---------- fila de prioridade (container/heap) ----------

type filaPrioridade []pedido

func (f filaPrioridade) Len() int { return len(f) }
func (f filaPrioridade) Less(i, j int) bool {
	if f[i].prioridade != f[j].prioridade {
		return f[i].prioridade > f[j].prioridade
	}
	return f[i].ordem < f[j].ordem
}
func (f filaPrioridade) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f *filaPrioridade) Push(x any)   { *f = append(*f, x.(pedido)) }
func (f *filaPrioridade) Pop() any {
	v := (*f)[len(*f)-1]
	*f = (*f)[:len(*f)-1]
	return v
}
