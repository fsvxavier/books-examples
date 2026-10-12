package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// resultado de um caso: ou tem nota, ou tem erro, nunca os dois.
type resultado struct {
	caso int
	nota float64
	err  error
}

var errJuizFora = errors.New("juiz indisponível")

// julga simula um juiz: leva um tempo e falha a cada quarto caso.
func julga(ctx context.Context, caso int) (float64, error) {
	select {
	case <-time.After(10 * time.Millisecond):
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	if caso%4 == 3 {
		return 0, errJuizFora
	}
	return float64(caso%10) / 10, nil
}

// executa avalia os casos com no máximo workers goroutines e devolve os
// resultados na ordem dos casos, sem mutex: cada posição é escrita por uma só goroutine.
func executa(ctx context.Context, casos []int, workers int) []resultado {
	saida := make([]resultado, len(casos))
	indices := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range indices {
				nota, err := julga(ctx, casos[i])
				saida[i] = resultado{caso: casos[i], nota: nota, err: err}
			}
		}()
	}
	for i := range casos {
		indices <- i
	}
	close(indices)
	wg.Wait() // só retorna quando todas as goroutines terminaram
	return saida
}

func main() {
	casos := make([]int, 12)
	for i := range casos {
		casos[i] = i
	}
	inicio := time.Now()
	res := executa(context.Background(), casos, 4)
	fmt.Println("12 casos com 4 workers terminaram antes dos 120ms do modo sequencial?", time.Since(inicio) < 120*time.Millisecond)

	soma, avaliados, falhas := 0.0, 0, 0
	for _, r := range res {
		if r.err != nil {
			falhas++
			fmt.Printf("caso %2d: sem nota (%v)\n", r.caso, r.err)
			continue
		}
		soma += r.nota
		avaliados++
	}
	// A média só considera quem foi avaliado: falha não é nota zero.
	fmt.Printf("avaliados=%d falhas=%d média=%.3f\n", avaliados, falhas, soma/float64(avaliados))
}
