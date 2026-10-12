package main

import (
	"fmt"
	"sync"
)

// Compartilhado e uma "referencia inteligente" (smart reference): abre o recurso pesado quando o
// primeiro cliente chega e o FECHA quando o ultimo cliente solta a referencia.
type Compartilhado[T interface{ Close() error }] struct {
	abrir func() (T, error)

	mu        sync.Mutex
	recurso   T
	refs      int
	Aberturas int
	Fechadas  int
}

func NovoCompartilhado[T interface{ Close() error }](abrir func() (T, error)) *Compartilhado[T] {
	return &Compartilhado[T]{abrir: abrir}
}

// Adquirir devolve o recurso e uma funcao que o devolve. Chamar a funcao duas vezes e inofensivo.
func (c *Compartilhado[T]) Adquirir() (T, func(), error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.refs == 0 {
		r, err := c.abrir()
		if err != nil {
			var zero T
			return zero, nil, fmt.Errorf("abrindo recurso: %w", err)
		}
		c.recurso = r
		c.Aberturas++
	}
	c.refs++
	var uma sync.Once
	soltar := func() {
		uma.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.refs--
			if c.refs == 0 {
				c.recurso.Close()
				c.Fechadas++
				var zero T
				c.recurso = zero
			}
		})
	}
	return c.recurso, soltar, nil
}

func (c *Compartilhado[T]) Clientes() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.refs
}
