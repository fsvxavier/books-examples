package main

import (
	"fmt"
	"sync"
	"time"
)

type entrada[V any] struct {
	valor    V
	expiraEm time.Time
}

// Cache guarda valores com prazo de validade (TTL). Seguro para uso concorrente.
type Cache[V any] struct {
	mu    sync.RWMutex
	itens map[string]entrada[V]
	ttl   time.Duration
}

func NovoCache[V any](ttl time.Duration) *Cache[V] {
	return &Cache[V]{itens: make(map[string]entrada[V]), ttl: ttl}
}

func (c *Cache[V]) Definir(chave string, valor V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.itens[chave] = entrada[V]{valor, time.Now().Add(c.ttl)}
}

// Obter devolve o valor se existir e ainda estiver válido.
func (c *Cache[V]) Obter(chave string) (V, bool) {
	c.mu.RLock()
	e, ok := c.itens[chave]
	c.mu.RUnlock()

	if !ok || time.Now().After(e.expiraEm) {
		var zero V
		return zero, false
	}
	return e.valor, true
}

// LimparExpirados remove as entradas vencidas e devolve quantas foram removidas.
func (c *Cache[V]) LimparExpirados() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	removidas := 0
	agora := time.Now()
	for k, e := range c.itens {
		if agora.After(e.expiraEm) {
			delete(c.itens, k)
			removidas++
		}
	}
	return removidas
}

func (c *Cache[V]) Tamanho() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.itens)
}

func main() {
	c := NovoCache[string](time.Minute)
	c.Definir("usuario:1", "Ana")
	v, ok := c.Obter("usuario:1")
	fmt.Println(v, ok) // Ana true
}
