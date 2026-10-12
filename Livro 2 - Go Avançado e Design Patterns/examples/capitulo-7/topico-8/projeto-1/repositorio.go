package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Repositorio e a interface do SERVICO; o proxy e o servico real a implementam.
type Repositorio interface {
	Buscar(ctx context.Context, id int) (string, error)
	Close() error
}

// ---------- O servico real: abrir e caro ----------

type repoReal struct{ id int }

var aberturas int
var mu sync.Mutex

func abrir(ctx context.Context) (Repositorio, error) {
	mu.Lock()
	aberturas++
	n := aberturas
	mu.Unlock()
	select {
	case <-time.After(100 * time.Millisecond): // handshake, autenticacao...
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return &repoReal{id: n}, nil
}

func (r *repoReal) Buscar(_ context.Context, id int) (string, error) {
	return fmt.Sprintf("registro %d (conexao %d)", id, r.id), nil
}
func (r *repoReal) Close() error { return nil }

// ---------- Proxy virtual: abre na PRIMEIRA chamada, e tenta de novo se falhar ----------

type ProxyPreguicoso struct {
	abrir func(context.Context) (Repositorio, error)

	mu   sync.Mutex
	real Repositorio
}

func NovoProxyPreguicoso(abrir func(context.Context) (Repositorio, error)) *ProxyPreguicoso {
	return &ProxyPreguicoso{abrir: abrir}
}

// obter devolve o servico real, criando-o se for preciso. Erros NAO ficam guardados.
func (p *ProxyPreguicoso) obter(ctx context.Context) (Repositorio, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.real != nil {
		return p.real, nil
	}
	r, err := p.abrir(ctx)
	if err != nil {
		return nil, fmt.Errorf("abrindo repositorio: %w", err)
	}
	p.real = r
	return r, nil
}

func (p *ProxyPreguicoso) Buscar(ctx context.Context, id int) (string, error) {
	r, err := p.obter(ctx)
	if err != nil {
		return "", err
	}
	return r.Buscar(ctx, id)
}

// Close so fecha se algum dia abriu: o proxy gerencia o ciclo de vida.
func (p *ProxyPreguicoso) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.real == nil {
		return nil
	}
	err := p.real.Close()
	p.real = nil
	return err
}

var ErrIndisponivel = errors.New("banco indisponivel")
