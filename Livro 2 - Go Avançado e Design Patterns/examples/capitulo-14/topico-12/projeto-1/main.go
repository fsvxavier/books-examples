package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var ErrTimeout = errors.New("dependência não respondeu")

func main() {
	const chamadores, porChamador = 4, 10
	fmt.Println("dependência fora do ar: cada chamada espera 100 ms e falha (40 chamadas, 4 em paralelo)")
	for _, comDisjuntor := range []bool{false, true} {
		var chamadasReais atomic.Int32
		dependencia := func(ctx context.Context) error {
			chamadasReais.Add(1)
			time.Sleep(100 * time.Millisecond)
			return ErrTimeout
		}
		d := &Disjuntor{LimiteFalhas: 4, Espera: time.Second}
		var wg sync.WaitGroup
		var rapidas, lentas atomic.Int32
		ini := time.Now()
		for range chamadores {
			wg.Go(func() {
				for range porChamador {
					t := time.Now()
					var err error
					if comDisjuntor {
						err = d.Executar(context.Background(), dependencia)
					} else {
						err = dependencia(context.Background())
					}
					_ = err
					if time.Since(t) < 10*time.Millisecond {
						rapidas.Add(1)
					} else {
						lentas.Add(1)
					}
				}
			})
		}
		wg.Wait()
		nome := "sem disjuntor"
		if comDisjuntor {
			nome = "com disjuntor"
		}
		fmt.Printf("  %-14s total %5v | chamadas à dependência %2d | respostas lentas %2d | respostas imediatas %2d | estado final: %v\n",
			nome, time.Since(ini).Round(10*time.Millisecond), chamadasReais.Load(), lentas.Load(), rapidas.Load(), d.Estado())
	}
}
