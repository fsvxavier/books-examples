package main

import (
	"context"
	"fmt"
	"sync"
)

// Func e o componente: uma funcao de busca. Cada decorador recebe uma Func e devolve outra Func.
type Func[A comparable, R any] func(ctx context.Context, a A) (R, error)

type Decorador[A comparable, R any] func(Func[A, R]) Func[A, R]

// Aplicar monta a pilha com o PRIMEIRO da lista mais por fora.
func Aplicar[A comparable, R any](f Func[A, R], ds ...Decorador[A, R]) Func[A, R] {
	for i := len(ds) - 1; i >= 0; i-- {
		f = ds[i](f)
	}
	return f
}

// Memoizar guarda os SUCESSOS por argumento. Erros nao sao guardados (podem ser temporarios).
func Memoizar[A comparable, R any]() Decorador[A, R] {
	return func(prox Func[A, R]) Func[A, R] {
		var (
			mu    sync.Mutex
			cache = map[A]R{}
		)
		return func(ctx context.Context, a A) (R, error) {
			mu.Lock()
			if v, ok := cache[a]; ok {
				mu.Unlock()
				return v, nil
			}
			mu.Unlock()
			v, err := prox(ctx, a)
			if err != nil {
				var zero R
				return zero, err
			}
			mu.Lock()
			cache[a] = v
			mu.Unlock()
			return v, nil
		}
	}
}

// Retentar repete ate n vezes enquanto houver erro e o contexto estiver vivo.
func Retentar[A comparable, R any](n int) Decorador[A, R] {
	return func(prox Func[A, R]) Func[A, R] {
		return func(ctx context.Context, a A) (R, error) {
			var (
				v   R
				err error
			)
			for i := 0; i < n; i++ {
				if ctx.Err() != nil {
					break
				}
				if v, err = prox(ctx, a); err == nil {
					return v, nil
				}
			}
			var zero R
			return zero, fmt.Errorf("depois de %d tentativas: %w", n, err)
		}
	}
}

// Anotar registra cada chamada que passa por ele.
func Anotar[A comparable, R any](nome string, registro func(string)) Decorador[A, R] {
	return func(prox Func[A, R]) Func[A, R] {
		return func(ctx context.Context, a A) (R, error) {
			v, err := prox(ctx, a)
			registro(fmt.Sprintf("%s(%v) -> err=%v", nome, a, err))
			return v, err
		}
	}
}
