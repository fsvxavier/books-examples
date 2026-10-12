package grafo

import (
	"context"
	"sync"
)

// Mapear aplica f a cada entrada, com no máximo limite chamadas ao mesmo tempo,
// e devolve os resultados **na ordem das entradas**, não na ordem de término.
// A primeira falha cancela as demais e é a que se devolve.
func Mapear[T, R any](ctx context.Context, entradas []T, limite int, f func(context.Context, T) (R, error)) ([]R, error) {
	ctx, cancelar := context.WithCancelCause(ctx)
	defer cancelar(nil)

	saidas := make([]R, len(entradas))
	vaga := make(chan struct{}, max(limite, 1))
	var (
		wg       sync.WaitGroup
		umaVez   sync.Once
		primeira error
	)
	for i, e := range entradas {
		wg.Go(func() {
			select {
			case vaga <- struct{}{}:
				defer func() { <-vaga }()
			case <-ctx.Done():
				return
			}
			r, err := f(ctx, e)
			if err != nil {
				umaVez.Do(func() { primeira = err; cancelar(err) })
				return
			}
			saidas[i] = r
		})
	}
	wg.Wait()
	if primeira != nil {
		return nil, primeira
	}
	return saidas, ctx.Err()
}
