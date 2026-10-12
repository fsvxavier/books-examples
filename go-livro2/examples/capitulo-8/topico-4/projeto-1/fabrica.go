package main

import (
	"fmt"
	"sync"
)

// Fabrica gerencia o pool de flyweights: dado o estado intrinseco (a chave K),
// devolve o flyweight existente ou cria um novo, UMA unica vez, mesmo com concorrencia.
type Fabrica[K comparable, V any] struct {
	criar func(K) (V, error)

	mu      sync.RWMutex
	pool    map[K]V
	Criados int // so para o experimento (protegido por mu)
}

func NovaFabrica[K comparable, V any](criar func(K) (V, error)) *Fabrica[K, V] {
	return &Fabrica[K, V]{criar: criar, pool: map[K]V{}}
}

func (f *Fabrica[K, V]) Obter(chave K) (V, error) {
	f.mu.RLock()
	v, ok := f.pool[chave]
	f.mu.RUnlock()
	if ok {
		return v, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if v, ok := f.pool[chave]; ok { // outra goroutine pode ter criado enquanto esperavamos a trava
		return v, nil
	}
	v, err := f.criar(chave)
	if err != nil {
		var zero V
		return zero, fmt.Errorf("criando flyweight %v: %w", chave, err)
	}
	f.pool[chave] = v
	f.Criados++
	return v, nil
}

func (f *Fabrica[K, V]) Tamanho() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.pool)
}
