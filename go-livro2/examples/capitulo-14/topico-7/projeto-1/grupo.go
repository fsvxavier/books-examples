package main

import (
	"context"

	"golang.org/x/sync/singleflight"
)

// Grupo é um singleflight.Group tipado: evita o any e a conversão no
// ponto de uso.
type Grupo[V any] struct{ g singleflight.Group }

// Do executa fn uma vez para a chave, enquanto houver uma execução em
// andamento. Quem chega durante a execução espera e recebe o MESMO
// resultado (compartilhado = true).
func (g *Grupo[V]) Do(chave string, fn func() (V, error)) (V, bool, error) {
	v, err, comp := g.g.Do(chave, func() (any, error) { return fn() })
	if v == nil {
		var zero V
		return zero, comp, err
	}
	return v.(V), comp, err
}

// DoCtx é a versão que cada chamador espera com o SEU contexto: o
// trabalho roda desligado do contexto de quem chegou primeiro (por isso
// o fn recebe um contexto próprio) e cada um desiste por conta própria.
func (g *Grupo[V]) DoCtx(ctx context.Context, chave string,
	fn func(ctx context.Context) (V, error)) (V, bool, error) {

	ch := g.g.DoChan(chave, func() (any, error) {
		// WithoutCancel mantém os VALORES do contexto (trace id) mas
		// não herda o cancelamento nem o prazo do primeiro chamador.
		return fn(context.WithoutCancel(ctx))
	})
	select {
	case r := <-ch:
		if r.Val == nil {
			var zero V
			return zero, r.Shared, r.Err
		}
		return r.Val.(V), r.Shared, r.Err
	case <-ctx.Done():
		var zero V
		return zero, false, ctx.Err()
	}
}

func (g *Grupo[V]) Esquecer(chave string) { g.g.Forget(chave) }
