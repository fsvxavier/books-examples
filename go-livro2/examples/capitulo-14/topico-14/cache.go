package main

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type entrada[V any] struct {
	valor    V
	expiraEm time.Time
}

// Cache carrega sob demanda, guarda por um TTL e colapsa as cargas
// simultâneas da mesma chave em uma só.
type Cache[V any] struct {
	TTL           time.Duration
	ServirVelho   bool // se a recarga falhar, devolve o valor expirado
	Agora         func() time.Time
	Carregar      func(ctx context.Context, chave string) (V, error)
	ProtegerCarga bool // false = sem singleflight (para comparar)

	mu      sync.RWMutex
	dados   map[string]entrada[V]
	grupo   Grupo[V]
	Cargas  atomic.Int64
	Acertos atomic.Int64
}

func (c *Cache[V]) agora() time.Time {
	if c.Agora != nil {
		return c.Agora()
	}
	return time.Now()
}

// Obter devolve o valor e se ele é velho (expirado, servido por falha).
func (c *Cache[V]) Obter(ctx context.Context, chave string) (V, bool, error) {
	c.mu.RLock()
	e, ok := c.dados[chave]
	c.mu.RUnlock()
	if ok && c.agora().Before(e.expiraEm) {
		c.Acertos.Add(1)
		return e.valor, false, nil
	}

	carregar := func(ctx context.Context) (V, error) {
		c.Cargas.Add(1)
		v, err := c.Carregar(ctx, chave)
		if err != nil {
			var zero V
			return zero, err // erro NÃO é guardado
		}
		c.mu.Lock()
		if c.dados == nil {
			c.dados = map[string]entrada[V]{}
		}
		c.dados[chave] = entrada[V]{v, c.agora().Add(c.TTL)}
		c.mu.Unlock()
		return v, nil
	}

	var v V
	var err error
	if c.ProtegerCarga {
		v, _, err = c.grupo.DoCtx(ctx, chave, carregar)
	} else {
		v, err = carregar(ctx)
	}
	if err != nil && ok && c.ServirVelho {
		return e.valor, true, nil
	}
	return v, false, err
}
