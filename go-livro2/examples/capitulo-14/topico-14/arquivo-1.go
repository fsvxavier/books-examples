package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type falso struct {
	chamadas atomic.Int32
	latencia time.Duration
	preco    int
	falha    atomic.Bool
}

func (f *falso) Preco(ctx context.Context, _ string) (int, error) {
	f.chamadas.Add(1)
	select {
	case <-time.After(f.latencia):
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	if f.falha.Load() {
		return 0, errors.New("erro 500")
	}
	return f.preco, nil
}

func provedor(nome string, f *falso, cota rate.Limit, burst int, timeout time.Duration) *Provedor {
	return &Provedor{
		Nome: nome, Preco: f.Preco, Timeout: timeout,
		Limite: rate.NewLimiter(cota, burst),
		Disj:   &Disjuntor{LimiteFalhas: 3, Espera: time.Minute},
	}
}

func main() {
	rapido := &falso{latencia: 20 * time.Millisecond, preco: 10500}
	medio := &falso{latencia: 60 * time.Millisecond, preco: 9900}
	lento := &falso{latencia: 2 * time.Second, preco: 9000} // o mais barato, mas lento demais
	c := NovoCotador(time.Minute, 300*time.Millisecond,
		provedor("rapido", rapido, 100, 10, 200*time.Millisecond),
		provedor("medio", medio, 100, 10, 200*time.Millisecond),
		provedor("lento", lento, 100, 10, 200*time.Millisecond),
	)

	fmt.Println("1) uma cotação: o provedor lento estoura o prazo, os outros respondem")
	ini := time.Now()
	q, err := c.Cotar(context.Background(), "notebook")
	fmt.Printf("   melhor: %s por %d | falhas: [%s] | err=%v | levou %v\n",
		q.Provedor, q.Preco, q.ResumoFalhas(), err, time.Since(ini).Round(10*time.Millisecond))

	fmt.Println("2) a mesma cotação de novo: vem do cache")
	ini = time.Now()
	_, _ = c.Cotar(context.Background(), "notebook")
	fmt.Printf("   levou %v | consultas reais ao provedor rapido: %d\n", time.Since(ini).Round(time.Millisecond), rapido.chamadas.Load())

	fmt.Println("3) 500 requisições simultâneas para um produto frio")
	var wg sync.WaitGroup
	for range 500 {
		wg.Go(func() { _, _ = c.Cotar(context.Background(), "monitor") })
	}
	wg.Wait()
	fmt.Printf("   cargas: %d (1 de notebook + 1 de monitor) | chamadas: rapido=%d medio=%d lento=%d\n",
		c.Cargas(), rapido.chamadas.Load(), medio.chamadas.Load(), lento.chamadas.Load())

	fmt.Println("4) um provedor cai: o disjuntor para de chamá-lo")
	medio.falha.Store(true)
	antes := medio.chamadas.Load()
	for i := range 6 {
		_, _ = c.Cotar(context.Background(), fmt.Sprint("produto-", i))
	}
	fmt.Printf("   seis cotações novas: o provedor medio recebeu %d chamadas (disjuntor: %v)\n",
		medio.chamadas.Load()-antes, c.Provedores[1].Disj.Estado())
}
