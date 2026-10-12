package main

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Três formas de limitar a N tarefas simultâneas.

func comCanal(n, tarefas int, f func()) {
	sem := make(chan struct{}, n)
	var wg sync.WaitGroup
	for range tarefas {
		sem <- struct{}{} // adquire (bloqueia se cheio)
		wg.Go(func() {
			defer func() { <-sem }() // libera
			f()
		})
	}
	wg.Wait()
}

func comWeighted(n int64, tarefas int, f func()) {
	sem := semaphore.NewWeighted(n)
	ctx := context.Background()
	var wg sync.WaitGroup
	for range tarefas {
		_ = sem.Acquire(ctx, 1)
		wg.Go(func() {
			defer sem.Release(1)
			f()
		})
	}
	wg.Wait()
}

func comErrgroup(n, tarefas int, f func()) {
	var g errgroup.Group
	g.SetLimit(n)
	for range tarefas {
		g.Go(func() error { f(); return nil })
	}
	_ = g.Wait()
}

func formas() {
	fmt.Println("== 1. três formas de limitar a 10 tarefas simultâneas (1000 tarefas de 1 ms) ==")
	for _, c := range []struct {
		nome string
		f    func(func())
	}{
		{"canal com buffer", func(f func()) { comCanal(10, 1000, f) }},
		{"semaphore.Weighted", func(f func()) { comWeighted(10, 1000, f) }},
		{"errgroup.SetLimit", func(f func()) { comErrgroup(10, 1000, f) }},
	} {
		ini := time.Now()
		c.f(func() { time.Sleep(time.Millisecond) })
		fmt.Printf("  %-20s %v\n", c.nome, time.Since(ini).Round(5*time.Millisecond))
	}

	fmt.Println("   custo por tarefa (tarefa vazia, limite 10):")
	for _, c := range []struct {
		nome string
		f    func(func())
	}{
		{"canal com buffer", func(f func()) { comCanal(10, 1000, f) }},
		{"semaphore.Weighted", func(f func()) { comWeighted(10, 1000, f) }},
		{"errgroup.SetLimit", func(f func()) { comErrgroup(10, 1000, f) }},
	} {
		r := testing.Benchmark(func(b *testing.B) {
			for b.Loop() {
				c.f(func() {})
			}
		})
		fmt.Printf("  %-20s %6.0f ns/tarefa  %4.1f allocs/tarefa\n", c.nome, float64(r.NsPerOp())/1000, float64(r.AllocsPerOp())/1000)
	}
}

func pesos() {
	fmt.Println("== 2. Weighted: orçamento de memória de 100 unidades, FIFO ==")
	sem := semaphore.NewWeighted(100)
	ctx := context.Background()
	ini := time.Now()
	var mu sync.Mutex
	type largada struct {
		ms   int64
		nome string
	}
	var largadas []largada
	marca := func(nome string) {
		mu.Lock()
		largadas = append(largadas, largada{time.Since(ini).Milliseconds() / 50 * 50, nome})
		mu.Unlock()
	}
	var wg sync.WaitGroup
	jobs := []struct {
		nome string
		peso int64
	}{{"A", 60}, {"B", 60}, {"C", 30}, {"D", 30}, {"E", 10}}
	for _, j := range jobs {
		time.Sleep(2 * time.Millisecond) // garante a ordem de chegada A, B, C, D, E
		wg.Go(func() {
			_ = sem.Acquire(ctx, j.peso)
			marca(fmt.Sprintf("%s (peso %d)", j.nome, j.peso))
			time.Sleep(100 * time.Millisecond)
			sem.Release(j.peso)
		})
	}
	wg.Wait()
	slices.SortFunc(largadas, func(a, b largada) int {
		return cmp.Or(cmp.Compare(a.ms, b.ms), cmp.Compare(a.nome, b.nome))
	})
	for _, l := range largadas {
		fmt.Printf("  +%3dms começa %s\n", l.ms, l.nome)
	}
	fmt.Println("  Com A rodando sobravam 40: C (30) e E (10) caberiam, mas ficaram atrás de B (60).")
	fmt.Println("  Depois, com B e C rodando sobravam 10: E (10) caberia, mas ficou atrás de D (30).")
	fmt.Println("  O semáforo é FIFO estrito: ninguém fura a fila, nem quando cabe.")

	fmt.Println("== 3. TryAcquire não espera; Acquire respeita o contexto ==")
	s2 := semaphore.NewWeighted(10)
	fmt.Println("  TryAcquire(8):", s2.TryAcquire(8))
	fmt.Println("  TryAcquire(8):", s2.TryAcquire(8), "(só restam 2)")
	ctxCurto, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()
	fmt.Println("  Acquire(8) com prazo de 30 ms:", s2.Acquire(ctxCurto, 8))
	ctxCurto2, cancel2 := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel2()
	fmt.Println("  Acquire(11) em um semáforo de 10:", s2.Acquire(ctxCurto2, 11), "(nunca caberia)")
	func() {
		defer func() { fmt.Println("  Release a mais:", recover()) }()
		s2.Release(100)
	}()
}

func main() {
	formas()
	pesos()
}
