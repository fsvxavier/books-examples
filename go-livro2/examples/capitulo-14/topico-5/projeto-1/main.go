package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

func colapso() {
	fmt.Println("== 1. cem chamadores, uma execução ==")
	var g Grupo[int]
	var execucoes atomic.Int32
	var wg sync.WaitGroup
	var compartilhados atomic.Int32
	for range 100 {
		wg.Go(func() {
			_, comp, _ := g.Do("preco:ABC", func() (int, error) {
				execucoes.Add(1)
				time.Sleep(50 * time.Millisecond) // a consulta lenta
				return 4200, nil
			})
			if comp {
				compartilhados.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Printf("execuções: %d | chamadores que receberam resultado compartilhado: %d\n", execucoes.Load(), compartilhados.Load())

	// Depois que terminou, uma chamada nova executa de novo: singleflight NÃO é cache.
	_, _, _ = g.Do("preco:ABC", func() (int, error) { execucoes.Add(1); return 4300, nil })
	fmt.Println("execuções depois de uma chamada nova:", execucoes.Load(), "(não guarda o resultado)")
}

func erroCompartilhado() {
	fmt.Println("== 2. o erro também é compartilhado ==")
	var g Grupo[string]
	var wg sync.WaitGroup
	var iguais atomic.Int32
	var primeiro atomic.Pointer[error]
	for range 10 {
		wg.Go(func() {
			_, _, err := g.Do("k", func() (string, error) {
				time.Sleep(20 * time.Millisecond)
				return "", errors.New("banco fora do ar")
			})
			if primeiro.CompareAndSwap(nil, &err) || *primeiro.Load() == err {
				iguais.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("chamadores com exatamente o mesmo erro:", iguais.Load(), "de 10")
}

func panicRepassado() {
	fmt.Println("== 3. um panic em fn chega a TODOS os que esperam ==")
	var g Grupo[int]
	var wg sync.WaitGroup
	var panicaram atomic.Int32
	for range 20 {
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					panicaram.Add(1)
				}
			}()
			_, _, _ = g.Do("k", func() (int, error) {
				time.Sleep(20 * time.Millisecond)
				var m map[string]int
				m["x"] = 1
				return 0, nil
			})
		})
	}
	wg.Wait()
	fmt.Println("chamadores que receberam o panic:", panicaram.Load(), "de 20")
}

func contextoDoPrimeiro() {
	fmt.Println("== 4. o contexto do primeiro chamador governa todos ==")
	trabalho := func(ctx context.Context) (string, error) {
		select {
		case <-time.After(150 * time.Millisecond):
			return "resultado", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Versão ingênua: usa o contexto de quem chegou primeiro (prazo curto).
	var g Grupo[string]
	var wg sync.WaitGroup
	res := make([]string, 4)
	for i := range 4 {
		wg.Go(func() {
			ctx := context.Background()
			if i == 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 30*time.Millisecond) // o primeiro, apressado
				defer cancel()
			} else {
				time.Sleep(5 * time.Millisecond) // chegam depois, sem pressa
			}
			v, _, err := g.Do("k", func() (string, error) { return trabalho(ctx) })
			res[i] = fmt.Sprintf("%q %v", v, err)
		})
	}
	wg.Wait()
	fmt.Println("ingênua:")
	for i, r := range res {
		fmt.Printf("  chamador %d (prazo %s): %s\n", i, map[bool]string{true: "30ms", false: "sem prazo"}[i == 0], r)
	}

	// Versão correta: DoCtx. O trabalho não depende do contexto de ninguém.
	var h Grupo[string]
	res = make([]string, 4)
	for i := range 4 {
		wg.Go(func() {
			ctx := context.Background()
			if i == 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 30*time.Millisecond)
				defer cancel()
			} else {
				time.Sleep(5 * time.Millisecond)
			}
			v, _, err := h.DoCtx(ctx, "k", trabalho)
			res[i] = fmt.Sprintf("%q %v", v, err)
		})
	}
	wg.Wait()
	fmt.Println("com DoCtx:")
	for i, r := range res {
		fmt.Printf("  chamador %d (prazo %s): %s\n", i, map[bool]string{true: "30ms", false: "sem prazo"}[i == 0], r)
	}
}

func resultadoMutavel() {
	fmt.Println("== 5. o resultado compartilhado é o MESMO slice ==")
	var g Grupo[[]int]
	var wg sync.WaitGroup
	resultados := make([][]int, 2)
	for i := range 2 {
		wg.Go(func() {
			v, _, _ := g.Do("lista", func() ([]int, error) {
				time.Sleep(20 * time.Millisecond)
				return []int{3, 1, 2}, nil
			})
			resultados[i] = v
		})
	}
	wg.Wait()
	fmt.Println("mesmo array por baixo:", &resultados[0][0] == &resultados[1][0])
	slices.Sort(resultados[0]) // o chamador 0 ordena "a sua" lista
	fmt.Println("lista do chamador 1 depois de o 0 ordenar:", resultados[1])
	fmt.Println("(a correção: cada chamador faz slices.Clone do resultado)")
}

func main() {
	colapso()
	erroCompartilhado()
	panicRepassado()
	contextoDoPrimeiro()
	resultadoMutavel()
}
