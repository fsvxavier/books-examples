package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

var ErrFalhou = errors.New("tarefa falhou")

func primeiroErroCancela() {
	fmt.Println("== 1. o primeiro erro cancela o contexto do grupo ==")
	g, ctx := errgroup.WithContext(context.Background())
	var viramCancelamento atomic.Int32
	for i := 1; i <= 5; i++ {
		g.Go(func() error {
			if i == 3 {
				time.Sleep(20 * time.Millisecond)
				return fmt.Errorf("tarefa %d: %w", i, ErrFalhou)
			}
			select {
			case <-time.After(2 * time.Second): // trabalho longo
				return nil
			case <-ctx.Done():
				viramCancelamento.Add(1)
				return ctx.Err() // todas devolvem context.Canceled depois
			}
		})
	}
	ini := time.Now()
	err := g.Wait()
	fmt.Printf("Wait devolveu em %v: %v\n", time.Since(ini).Round(10*time.Millisecond), err)
	fmt.Println("tarefas que viram o cancelamento:", viramCancelamento.Load())
	fmt.Println("Cause(ctx):", context.Cause(ctx))
}

func soOPrimeiroErro() {
	fmt.Println("== 2. Wait devolve só o PRIMEIRO erro ==")
	var g errgroup.Group
	for i := 1; i <= 4; i++ {
		g.Go(func() error {
			time.Sleep(time.Duration(i) * 10 * time.Millisecond)
			return fmt.Errorf("erro da tarefa %d", i)
		})
	}
	fmt.Println("errgroup:", g.Wait())

	// Para guardar TODOS os erros: um coletor com errors.Join.
	var wg sync.WaitGroup
	var mu sync.Mutex
	var erros []error
	for i := 1; i <= 4; i++ {
		wg.Go(func() { // sync.WaitGroup.Go existe desde o Go 1.25
			time.Sleep(time.Duration(i) * 10 * time.Millisecond)
			mu.Lock()
			erros = append(erros, fmt.Errorf("erro da tarefa %d", i))
			mu.Unlock()
		})
	}
	wg.Wait()
	fmt.Println("errors.Join, todos os erros:")
	fmt.Println(errors.Join(erros...))
}

func contextoMorreDepoisDoWait() {
	fmt.Println("== 3. o contexto do grupo é cancelado quando Wait retorna ==")
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error { return nil }) // nenhum erro
	_ = g.Wait()
	fmt.Println("ctx.Err() depois de um Wait SEM erro:", ctx.Err())
	fmt.Println("Cause:", context.Cause(ctx))
}

func limite() {
	fmt.Println("== 4. SetLimit e TryGo ==")
	var g errgroup.Group
	g.SetLimit(3)
	var ativas, maxAtivas atomic.Int32
	ini := time.Now()
	for range 12 {
		g.Go(func() error { // Go BLOQUEIA quando o limite está cheio
			n := ativas.Add(1)
			for {
				m := maxAtivas.Load()
				if n <= m || maxAtivas.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			ativas.Add(-1)
			return nil
		})
	}
	_ = g.Wait()
	fmt.Printf("12 tarefas de 50 ms com limite 3: máximo simultâneo %d, total %v\n",
		maxAtivas.Load(), time.Since(ini).Round(10*time.Millisecond))

	var h errgroup.Group
	h.SetLimit(1)
	parar := make(chan struct{})
	fmt.Println("TryGo com vaga:", h.TryGo(func() error { <-parar; return nil }))
	fmt.Println("TryGo sem vaga:", h.TryGo(func() error { return nil }))
	func() {
		defer func() { fmt.Println("SetLimit com tarefa ativa:", recover()) }()
		h.SetLimit(5)
	}()
	close(parar)
	_ = h.Wait()
}

func main() {
	primeiroErroCancela()
	soOPrimeiroErro()
	contextoMorreDepoisDoWait()
	limite()
}
