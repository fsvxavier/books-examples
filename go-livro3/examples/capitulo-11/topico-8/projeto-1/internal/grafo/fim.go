// Package grafo é um motor mínimo de grafos de estado para agentes.
//
// Um grafo tem nós (funções que recebem o estado e devolvem o novo estado),
// arestas fixas e arestas condicionais (funções que olham o estado e escolhem
// o próximo nó). O estado é passado por valor: cada nó devolve uma cópia nova.
package grafo

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Fim é o nome reservado do destino que encerra a execução.
const Fim = "fim"

// Erros do motor.
var (
	ErrLimite         = errors.New("limite de passos atingido")
	ErrNoDesconhecido = errors.New("nó desconhecido")
	ErrSemSaida       = errors.New("nó sem aresta de saída")
)

// No é a unidade de trabalho: um agente, uma ferramenta, uma decisão.
type No[S any] func(ctx context.Context, s S) (S, error)

// Rota escolhe o próximo nó olhando o estado.
type Rota[S any] func(s S) string

// Evento descreve um passo concluído; serve para log, métricas e testes.
type Evento struct {
	Passo   int
	No      string
	Proximo string
	Duracao time.Duration
	Err     error
}

// Grafo reúne nós e arestas. Monte-o antes de executar e não o altere depois.
type Grafo[S any] struct {
	inicio   string
	nos      map[string]No[S]
	saidas   map[string]Rota[S]
	Limite   int          // máximo de passos; protege contra ciclos que não terminam
	Observar func(Evento) // chamado depois de cada passo, se não for nil
}

// Novo cria um grafo que começa em inicio.
func Novo[S any](inicio string) *Grafo[S] {
	return &Grafo[S]{
		inicio: inicio,
		nos:    map[string]No[S]{},
		saidas: map[string]Rota[S]{},
		Limite: 25,
	}
}

// No registra um nó.
func (g *Grafo[S]) No(nome string, f No[S]) *Grafo[S] {
	g.nos[nome] = f
	return g
}

// Aresta liga de a para.
func (g *Grafo[S]) Aresta(de, para string) *Grafo[S] {
	g.saidas[de] = func(S) string { return para }
	return g
}

// Condicional liga de ao nó que a rota escolher.
func (g *Grafo[S]) Condicional(de string, r Rota[S]) *Grafo[S] {
	g.saidas[de] = r
	return g
}

// Validar confere o que dá para conferir sem executar: início, nós sem saída.
func (g *Grafo[S]) Validar() error {
	if _, ok := g.nos[g.inicio]; !ok {
		return fmt.Errorf("início %q: %w", g.inicio, ErrNoDesconhecido)
	}
	for nome := range g.nos {
		if _, ok := g.saidas[nome]; !ok {
			return fmt.Errorf("nó %q: %w", nome, ErrSemSaida)
		}
	}
	return nil
}

// Ponto guarda e recupera o ponto de execução, para retomar depois de uma falha.
type Ponto[S any] interface {
	// Guardar registra o próximo nó a executar e o estado atual.
	Guardar(proximo string, s S) error
	// Carregar devolve o último ponto salvo; ok é false se não houver.
	Carregar() (proximo string, s S, ok bool, err error)
}

// Executar roda o grafo até Fim, até um erro ou até o contexto acabar.
func (g *Grafo[S]) Executar(ctx context.Context, s S) (S, error) {
	return g.ExecutarComPonto(ctx, s, nil)
}

// ExecutarComPonto é como Executar, mas grava um ponto depois de cada nó e,
// se já houver um ponto salvo, retoma dele em vez de começar do início.
func (g *Grafo[S]) ExecutarComPonto(ctx context.Context, s S, ponto Ponto[S]) (S, error) {
	atual := g.inicio
	if ponto != nil {
		proximo, salvo, ok, err := ponto.Carregar()
		if err != nil {
			return s, fmt.Errorf("carregar ponto: %w", err)
		}
		if ok {
			atual, s = proximo, salvo
		}
	}
	for passo := 1; atual != Fim; passo++ {
		if passo > g.Limite {
			return s, fmt.Errorf("%w (%d), parado em %q", ErrLimite, g.Limite, atual)
		}
		if err := ctx.Err(); err != nil {
			return s, err
		}
		no, ok := g.nos[atual]
		if !ok {
			return s, fmt.Errorf("%q: %w", atual, ErrNoDesconhecido)
		}
		inicio := time.Now()
		novo, err := no(ctx, s)
		ev := Evento{Passo: passo, No: atual, Duracao: time.Since(inicio), Err: err}
		if err != nil {
			g.notificar(ev)
			return s, fmt.Errorf("nó %q: %w", atual, err)
		}
		s = novo
		rota, ok := g.saidas[atual]
		if !ok {
			return s, fmt.Errorf("%q: %w", atual, ErrSemSaida)
		}
		ev.Proximo = rota(s)
		g.notificar(ev)
		atual = ev.Proximo
		if ponto != nil {
			if err := ponto.Guardar(atual, s); err != nil {
				return s, fmt.Errorf("guardar ponto: %w", err)
			}
		}
	}
	return s, nil
}

func (g *Grafo[S]) notificar(e Evento) {
	if g.Observar != nil {
		g.Observar(e)
	}
}
