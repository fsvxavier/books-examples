package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

type Pedido struct{ ID int }
type Baixado struct{ ID, Bytes int }
type Pronto struct{ ID, Soma int }

func ids(n int) []Pedido {
	p := make([]Pedido, n)
	for i := range p {
		p[i] = Pedido{i}
	}
	return p
}

func rodar(ctx context.Context, n int, workersBaixar, workersProcessar, buffer int,
	falhaEm int, sink func(Pronto) error) (processados int, err error) {

	g, ctx := errgroup.WithContext(ctx)
	var baixados atomic.Int32
	origem := Gerar(ctx, g, buffer, ids(n))
	baixado := Estagio(ctx, g, origem, workersBaixar, buffer, func(ctx context.Context, p Pedido) (Baixado, error) {
		if p.ID == falhaEm {
			return Baixado{}, errors.New("download falhou")
		}
		time.Sleep(5 * time.Millisecond) // I/O de rede
		baixados.Add(1)
		return Baixado{p.ID, p.ID * 10}, nil
	})
	pronto := Estagio(ctx, g, baixado, workersProcessar, buffer, func(ctx context.Context, b Baixado) (Pronto, error) {
		time.Sleep(2 * time.Millisecond) // CPU simulada
		return Pronto{b.ID, b.Bytes + 1}, nil
	})
	var total atomic.Int32
	Consumir(ctx, g, pronto, func(p Pronto) error {
		total.Add(1)
		return sink(p)
	})
	err = g.Wait()
	return int(total.Load()), err
}

func main() {
	fmt.Println("== 1. paralelismo por estágio: 200 pedidos (baixar 5 ms, processar 2 ms) ==")
	for _, w := range []int{1, 2, 4, 8, 16} {
		ini := time.Now()
		var vistos []int
		n, err := rodar(context.Background(), 200, w, max(1, w/2), 0, -1, func(p Pronto) error {
			vistos = append(vistos, p.ID)
			return nil
		})
		slices.Sort(vistos)
		ok := n == 200 && err == nil && slices.Equal(vistos, func() []int {
			s := make([]int, 200)
			for i := range s {
				s[i] = i
			}
			return s
		}())
		fmt.Printf("  baixar x%-2d processar x%-2d  %6v  (completo e sem duplicatas: %v)\n",
			w, max(1, w/2), time.Since(ini).Round(10*time.Millisecond), ok)
	}

	fmt.Println("== 2. um estágio falha: tudo para e nada vaza ==")
	antes := runtime.NumGoroutine()
	ini := time.Now()
	n, err := rodar(context.Background(), 10000, 4, 2, 0, 50, func(Pronto) error { return nil })
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("  erro: %v | gravados antes da falha: %d de 10000 | demorou %v\n", err, n, time.Since(ini).Round(10*time.Millisecond))
	fmt.Printf("  goroutines: %d antes, %d depois\n", antes, runtime.NumGoroutine())

	fmt.Println("== 3. contrapressão: consumidor lento (10 ms) e produtor rápido ==")
	for _, buf := range []int{0, 10, 100} {
		var gerados, gravados, maxVoo atomic.Int32
		g, ctx := errgroup.WithContext(context.Background())
		origem := make(chan int, buf)
		g.Go(func() error {
			defer close(origem)
			for i := range 300 {
				select {
				case origem <- i:
					n := gerados.Add(1)
					if v := n - gravados.Load(); v > maxVoo.Load() {
						maxVoo.Store(v)
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
		Consumir(ctx, g, origem, func(int) error {
			time.Sleep(10 * time.Millisecond)
			gravados.Add(1)
			return nil
		})
		_ = g.Wait()
		fmt.Printf("  buffer %-3d: no máximo %3d itens gerados e ainda não gravados (de 300)\n", buf, maxVoo.Load())
	}
}
