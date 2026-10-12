package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func debandada(proteger bool, goroutines int) (cargas int64, dur time.Duration) {
	c := &Cache[int]{
		TTL:           time.Minute,
		ProtegerCarga: proteger,
		Carregar: func(ctx context.Context, k string) (int, error) {
			time.Sleep(20 * time.Millisecond) // o banco lento
			return 42, nil
		},
	}
	ini := time.Now()
	var wg sync.WaitGroup
	largada := make(chan struct{})
	for range goroutines {
		wg.Go(func() {
			<-largada
			_, _, _ = c.Obter(context.Background(), "popular")
		})
	}
	close(largada)
	wg.Wait()
	return c.Cargas.Load(), time.Since(ini)
}

func main() {
	fmt.Println("chave fria, N requisições simultâneas, banco de 20 ms por consulta:")
	for _, n := range []int{10, 100, 1000} {
		sem, d1 := debandada(false, n)
		com, d2 := debandada(true, n)
		fmt.Printf("  N=%-5d sem proteção: %4d consultas (%v) | com singleflight: %d consulta (%v)\n",
			n, sem, d1.Round(time.Millisecond), com, d2.Round(time.Millisecond))
	}
}
