package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

// Provedor é uma fonte externa de preços, protegida por três mecanismos:
// um limitador de taxa (respeitar a cota dele), um disjuntor (parar de
// insistir quando está fora do ar) e um prazo por chamada.
type Provedor struct {
	Nome    string
	Preco   func(ctx context.Context, produto string) (int, error)
	Limite  *rate.Limiter
	Disj    *Disjuntor
	Timeout time.Duration
}

func (p *Provedor) consultar(ctx context.Context, produto string) (int, error) {
	if err := p.Limite.Wait(ctx); err != nil { // espera a vez, ou desiste com o prazo
		return 0, fmt.Errorf("limite: %w", err)
	}
	var preco int
	err := p.Disj.Executar(ctx, func(ctx context.Context) error {
		cctx, cancel := context.WithTimeout(ctx, p.Timeout)
		defer cancel()
		var err error
		preco, err = p.Preco(cctx, produto)
		return err
	})
	return preco, err
}

// Cotacao é o melhor preço entre os provedores que responderam.
type Cotacao struct {
	Provedor string
	Preco    int
	Falhas   map[string]error
}

func (c Cotacao) ResumoFalhas() string {
	var nomes []string
	for n := range c.Falhas {
		nomes = append(nomes, n)
	}
	sort.Strings(nomes)
	return strings.Join(nomes, ",")
}

var ErrTodosFalharam = errors.New("nenhum provedor respondeu")

// Cotador consulta os provedores em paralelo, tolera falhas parciais e
// guarda o resultado em um cache que colapsa consultas simultâneas.
type Cotador struct {
	Provedores []*Provedor
	Prazo      time.Duration // prazo da cotação inteira
	cache      *Cache[Cotacao]
}

func NovoCotador(ttl, prazo time.Duration, provedores ...*Provedor) *Cotador {
	c := &Cotador{Provedores: provedores, Prazo: prazo}
	c.cache = &Cache[Cotacao]{TTL: ttl, ProtegerCarga: true, Carregar: c.consultarTodos}
	return c
}

func (c *Cotador) Cotar(ctx context.Context, produto string) (Cotacao, error) {
	v, _, err := c.cache.Obter(ctx, produto)
	return v, err
}

func (c *Cotador) Cargas() int64 { return c.cache.Cargas.Load() }

// consultarTodos roda desligado do contexto do primeiro chamador (o
// singleflight) e por isso define o SEU prazo.
func (c *Cotador) consultarTodos(ctx context.Context, produto string) (Cotacao, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Prazo)
	defer cancel()

	type resp struct {
		preco int
		err   error
	}
	resps := make([]resp, len(c.Provedores))

	// SEM WithContext: uma falha de um provedor NÃO deve cancelar os outros.
	var g errgroup.Group
	for i, p := range c.Provedores {
		g.Go(func() error {
			preco, err := p.consultar(ctx, produto)
			resps[i] = resp{preco, err} // cada goroutine escreve só no seu índice
			return nil
		})
	}
	_ = g.Wait()

	melhor := Cotacao{Preco: -1, Falhas: map[string]error{}}
	for i, r := range resps {
		nome := c.Provedores[i].Nome
		switch {
		case r.err != nil:
			melhor.Falhas[nome] = r.err
		case melhor.Preco < 0 || r.preco < melhor.Preco:
			melhor.Provedor, melhor.Preco = nome, r.preco
		}
	}
	if melhor.Preco < 0 {
		return Cotacao{}, fmt.Errorf("%w: %v", ErrTodosFalharam, errors.Join(valores(melhor.Falhas)...))
	}
	return melhor, nil
}

func valores(m map[string]error) []error {
	var nomes []string
	for n := range m {
		nomes = append(nomes, n)
	}
	sort.Strings(nomes)
	out := make([]error, 0, len(m))
	for _, n := range nomes {
		out = append(out, fmt.Errorf("%s: %w", n, m[n]))
	}
	return out
}
