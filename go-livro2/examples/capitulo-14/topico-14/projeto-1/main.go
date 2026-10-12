package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

// main faz uma cotação com dois provedores e repete a consulta: a segunda vem
// do cache, sem chamar os provedores de novo.
func main() {
	novo := func(nome string, preco int, latencia time.Duration) *Provedor {
		return &Provedor{
			Nome: nome,
			Preco: func(ctx context.Context, _ string) (int, error) {
				select {
				case <-time.After(latencia):
					return preco, nil
				case <-ctx.Done():
					return 0, ctx.Err()
				}
			},
			Limite:  rate.NewLimiter(rate.Limit(100), 10),
			Disj:    &Disjuntor{LimiteFalhas: 3, Espera: time.Minute},
			Timeout: 100 * time.Millisecond,
		}
	}
	c := NovoCotador(time.Minute, 500*time.Millisecond, novo("rapido", 10000, 5*time.Millisecond), novo("medio", 9900, 20*time.Millisecond))
	for i := 1; i <= 2; i++ {
		inicio := time.Now()
		cot, err := c.Cotar(context.Background(), "notebook")
		fmt.Printf("cotação %d: %+v err=%v levou %v | cargas=%d\n", i, cot, err, time.Since(inicio).Round(time.Millisecond), c.Cargas())
	}
}
