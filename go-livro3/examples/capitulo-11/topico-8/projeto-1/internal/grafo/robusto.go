package grafo

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Tentar repete o nó até vezes tentativas quando o erro for transitório,
// esperando espera, 2*espera, 4*espera... entre elas. O contexto interrompe a espera.
func Tentar[S any](vezes int, espera time.Duration, transitorio func(error) bool, no No[S]) No[S] {
	return func(ctx context.Context, s S) (S, error) {
		var err error
		for i := 0; i < vezes; i++ {
			var novo S
			novo, err = no(ctx, s)
			if err == nil {
				return novo, nil
			}
			if !transitorio(err) || i == vezes-1 {
				break
			}
			select {
			case <-time.After(espera << i):
			case <-ctx.Done():
				return s, ctx.Err()
			}
		}
		return s, fmt.Errorf("depois de %d tentativas: %w", vezes, err)
	}
}

// ErrOrcamento indica que o gasto passou do permitido.
var ErrOrcamento = errors.New("orçamento esgotado")

// ComOrcamento recusa executar o nó se gasto(s) já alcançou o teto.
func ComOrcamento[S any](teto int, gasto func(S) int, no No[S]) No[S] {
	return func(ctx context.Context, s S) (S, error) {
		if g := gasto(s); g >= teto {
			return s, fmt.Errorf("gasto %d, teto %d: %w", g, teto, ErrOrcamento)
		}
		return no(ctx, s)
	}
}
