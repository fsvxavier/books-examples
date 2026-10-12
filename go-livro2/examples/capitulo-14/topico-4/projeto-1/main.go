package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

func main() {
	ctx := context.Background()
	itens := []int{1, 2, 3, 4, 5, 6, 7, 8}

	r, err := Mapear(ctx, itens, 3, func(ctx context.Context, i, n int) (string, error) {
		time.Sleep(time.Duration(10*(8-n)) * time.Millisecond) // terminam em ordem inversa
		return strings.Repeat("*", n), nil
	})
	fmt.Println("resultados em ordem:", r, err)

	var iniciadas atomic.Int32
	_, err = Mapear(ctx, make([]int, 100), 2, func(ctx context.Context, i, _ int) (int, error) {
		iniciadas.Add(1)
		if i == 3 {
			return 0, fmt.Errorf("item %d: %w", i, ErroDeNegocio)
		}
		select {
		case <-time.After(30 * time.Millisecond):
			return i, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	})
	fmt.Printf("erro: %v | tarefas iniciadas: %d de 100\n", err, iniciadas.Load())

	_, err = Mapear(ctx, itens, 2, func(ctx context.Context, i, _ int) (int, error) {
		if i == 2 {
			var m map[string]int
			m["x"] = 1
		}
		return i, nil
	})
	var ep *ErroPanic
	fmt.Println("panic virou erro:", errors.As(err, &ep), "|", err)
}
