package main

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// banco simula uma gravação: 2 ms fixos de ida e volta + 20 us por item.
func gravacao(itens int) { time.Sleep(2*time.Millisecond + time.Duration(itens)*20*time.Microsecond) }

func percentil(d []time.Duration, p float64) time.Duration {
	slices.Sort(d)
	return d[int(float64(len(d)-1)*p)]
}

func medir(nome string, produtores int, enviar func(i int) error, chamadas *atomic.Int32) {
	lat := make([]time.Duration, produtores)
	var wg sync.WaitGroup
	ini := time.Now()
	for i := range produtores {
		wg.Go(func() {
			t := time.Now()
			_ = enviar(i)
			lat[i] = time.Since(t)
		})
	}
	wg.Wait()
	fmt.Printf("  %-34s total %6v | gravações %4d | latência p50 %6v p99 %6v\n", nome,
		time.Since(ini).Round(time.Millisecond), chamadas.Load(),
		percentil(lat, 0.5).Round(time.Millisecond), percentil(lat, 0.99).Round(time.Millisecond))
}

func main() {
	const produtores = 1000
	fmt.Println("1000 produtores, cada um grava 1 item; o banco aceita 10 gravações ao mesmo tempo:")

	// Sem lote: uma gravação por item, com no máximo 10 conexões.
	var c1 atomic.Int32
	conexoes := make(chan struct{}, 10)
	medir("um a um (10 conexões)", produtores, func(int) error {
		conexoes <- struct{}{}
		defer func() { <-conexoes }()
		c1.Add(1)
		gravacao(1)
		return nil
	}, &c1)

	for _, cfg := range []struct {
		max    int
		espera time.Duration
	}{{10, 5 * time.Millisecond}, {100, 5 * time.Millisecond}, {500, 5 * time.Millisecond}} {
		var c atomic.Int32
		l := &Lote[int]{MaxItens: cfg.max, MaxEspera: cfg.espera, Gravar: func(_ context.Context, it []int) error {
			c.Add(1)
			gravacao(len(it))
			return nil
		}}
		medir(fmt.Sprintf("em lote (máx %d, espera %v)", cfg.max, cfg.espera), produtores, func(i int) error {
			return l.Enviar(context.Background(), i)
		}, &c)
		l.Fechar()
	}

	fmt.Println("Produtores lentos: 1 item a cada 20 ms; a espera máxima vira a latência:")
	for _, espera := range []time.Duration{2 * time.Millisecond, 20 * time.Millisecond, 100 * time.Millisecond} {
		var c atomic.Int32
		l := &Lote[int]{MaxItens: 100, MaxEspera: espera, Gravar: func(_ context.Context, it []int) error {
			c.Add(1)
			gravacao(len(it))
			return nil
		}}
		var lats []time.Duration
		var mu sync.Mutex
		var wg sync.WaitGroup
		for range 20 {
			time.Sleep(20 * time.Millisecond)
			wg.Go(func() {
				t := time.Now()
				_ = l.Enviar(context.Background(), 1)
				mu.Lock()
				lats = append(lats, time.Since(t))
				mu.Unlock()
			})
		}
		wg.Wait()
		l.Fechar()
		fmt.Printf("  MaxEspera %-6v latência mediana %6v | gravações %2d para 20 itens\n", espera, percentil(lats, 0.5).Round(time.Millisecond), c.Load())
	}
}
