// Package e11 tem um cache genérico com validade e um mapeamento paralelo genérico.
package e11

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type entrada[V any] struct {
	valor V
	vence time.Time
}

// Cache guarda valores de tipo V por chave de tipo K, cada um com um prazo de validade.
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	itens    map[K]entrada[V]
	validade time.Duration
	agora    func() time.Time
}

// NovoCache cria um cache cujos valores valem por validade. agora, se nulo, usa time.Now.
func NovoCache[K comparable, V any](validade time.Duration, agora func() time.Time) *Cache[K, V] {
	if agora == nil {
		agora = time.Now
	}
	return &Cache[K, V]{itens: make(map[K]entrada[V]), validade: validade, agora: agora}
}

// Guardar grava o valor com a validade padrão do cache.
func (c *Cache[K, V]) Guardar(chave K, valor V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.itens[chave] = entrada[V]{valor: valor, vence: c.agora().Add(c.validade)}
}

// Obter devolve o valor, se existir e não tiver vencido.
func (c *Cache[K, V]) Obter(chave K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.itens[chave]
	if !ok || !c.agora().Before(e.vence) {
		delete(c.itens, chave)
		var zero V
		return zero, false
	}
	return e.valor, true
}

// ObterOuCarregar devolve o valor guardado ou chama carregar, guarda o resultado e o devolve.
// Se carregar falhar, nada é guardado.
func (c *Cache[K, V]) ObterOuCarregar(chave K, carregar func() (V, error)) (V, error) {
	if v, ok := c.Obter(chave); ok {
		return v, nil
	}
	v, err := carregar()
	if err != nil {
		var zero V
		return zero, err
	}
	c.Guardar(chave, v)
	return v, nil
}

// MapearParalelo aplica f a cada entrada, com no máximo limite chamadas ao mesmo tempo.
// O resultado respeita a ordem da entrada. O primeiro erro cancela as chamadas pendentes.
func MapearParalelo[E, S any](ctx context.Context, entradas []E, limite int, f func(context.Context, E) (S, error)) ([]S, error) {
	saidas := make([]S, len(entradas))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(limite)
	for i, e := range entradas {
		if ctx.Err() != nil {
			break // já falhou (ou foi cancelado): não adianta despachar o resto
		}
		g.Go(func() error {
			s, err := f(ctx, e)
			if err != nil {
				return err
			}
			saidas[i] = s // cada goroutine escreve no seu próprio índice
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return saidas, nil
}
