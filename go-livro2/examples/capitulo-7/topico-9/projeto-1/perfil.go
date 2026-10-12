package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Perfil struct {
	ID    int    `json:"id"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
}

var (
	ErrNaoEncontrado = errors.New("perfil nao encontrado")
	ErrProibido      = errors.New("acesso proibido")
)

// Perfis e a interface do SERVICO. O servico real, os proxies e o cliente remoto a implementam.
type Perfis interface {
	Obter(ctx context.Context, id int) (Perfil, error)
	Atualizar(ctx context.Context, p Perfil) error
}

// ---------- Servico real: um "banco" em memoria com latencia ----------

type Banco struct {
	mu       sync.RWMutex
	dados    map[int]Perfil
	Latencia time.Duration
	Leituras atomic.Int64 // quantas vezes o banco REAL foi consultado
}

func NovoBanco(ps ...Perfil) *Banco {
	b := &Banco{dados: map[int]Perfil{}}
	for _, p := range ps {
		b.dados[p.ID] = p
	}
	return b
}

func (b *Banco) Obter(ctx context.Context, id int) (Perfil, error) {
	b.Leituras.Add(1)
	if b.Latencia > 0 {
		select {
		case <-time.After(b.Latencia):
		case <-ctx.Done():
			return Perfil{}, ctx.Err()
		}
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	p, ok := b.dados[id]
	if !ok {
		return Perfil{}, ErrNaoEncontrado
	}
	return p, nil
}

func (b *Banco) Atualizar(_ context.Context, p Perfil) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.dados[p.ID]; !ok {
		return ErrNaoEncontrado
	}
	b.dados[p.ID] = p
	return nil
}
