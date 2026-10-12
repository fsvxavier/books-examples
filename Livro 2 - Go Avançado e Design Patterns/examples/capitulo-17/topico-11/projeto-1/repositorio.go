// Package e09 mostra por que um repositório genérico não serve ao encurtador.
package e09

import (
	"context"
	"errors"
	"sync"
)

// Repositorio é o repositório genérico "de manual": qualquer entidade T com chave ID.
type Repositorio[T any, ID comparable] interface {
	Salvar(ctx context.Context, id ID, valor T) error
	Buscar(ctx context.Context, id ID) (T, error)
	Remover(ctx context.Context, id ID) error
}

// ErrNaoEncontrado indica uma chave ausente.
var ErrNaoEncontrado = errors.New("não encontrado")

// Memoria implementa Repositorio em memória.
type Memoria[T any, ID comparable] struct {
	mu    sync.Mutex
	dados map[ID]T
	// falhar, se não nulo, é chamado antes de cada Salvar e pode recusá-lo (para o teste).
	falhar func() error
}

// NovaMemoria cria um repositório em memória.
func NovaMemoria[T any, ID comparable]() *Memoria[T, ID] {
	return &Memoria[T, ID]{dados: make(map[ID]T)}
}

// Salvar grava o valor.
func (m *Memoria[T, ID]) Salvar(_ context.Context, id ID, v T) error {
	if m.falhar != nil {
		if err := m.falhar(); err != nil {
			return err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dados[id] = v
	return nil
}

// Buscar devolve o valor ou ErrNaoEncontrado.
func (m *Memoria[T, ID]) Buscar(_ context.Context, id ID) (T, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.dados[id]
	if !ok {
		var zero T
		return zero, ErrNaoEncontrado
	}
	return v, nil
}

// Remover apaga a chave.
func (m *Memoria[T, ID]) Remover(_ context.Context, id ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.dados, id)
	return nil
}

// Tamanho devolve quantas chaves há.
func (m *Memoria[T, ID]) Tamanho() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.dados)
}
