package main

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Gerar emite os itens e fecha o canal. Respeita o cancelamento.
func Gerar[E any](ctx context.Context, g *errgroup.Group, buffer int, itens []E) <-chan E {
	out := make(chan E, buffer)
	g.Go(func() error {
		defer close(out)
		for _, it := range itens {
			select {
			case out <- it:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})
	return out
}

// Estagio aplica f com n workers. Cada worker roda no grupo: o primeiro
// erro cancela o contexto e todos saem. O canal de saída é fechado quando
// o último worker termina.
func Estagio[E, S any](ctx context.Context, g *errgroup.Group, in <-chan E, n, buffer int,
	f func(context.Context, E) (S, error)) <-chan S {

	out := make(chan S, buffer)
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			for e := range in {
				s, err := f(ctx, e)
				if err != nil {
					return err
				}
				select {
				case out <- s:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}
	g.Go(func() error { // o fechador: espera os workers e fecha a saída
		wg.Wait()
		close(out)
		return nil
	})
	return out
}

// Consumir lê tudo, chamando f para cada item (um só consumidor).
func Consumir[E any](ctx context.Context, g *errgroup.Group, in <-chan E, f func(E) error) {
	g.Go(func() error {
		for e := range in {
			if err := f(e); err != nil {
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		return nil
	})
}
