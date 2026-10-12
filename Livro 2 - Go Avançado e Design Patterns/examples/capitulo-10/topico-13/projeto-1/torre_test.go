package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPrioridadeEDesempate(t *testing.T) {
	torre := NovaTorre()
	defer torre.Encerrar()
	ctx := context.Background()

	if err := torre.Solicitar(ctx, "A", 1); err != nil { // A ocupa a pista
		t.Fatal(err)
	}
	var mu sync.Mutex
	var ordem []string
	var wg sync.WaitGroup
	for i, p := range []struct {
		id   string
		prio int
	}{{"B", 1}, {"C", 5}, {"D", 1}, {"E", 5}} {
		wg.Go(func() {
			if err := torre.Solicitar(ctx, p.id, p.prio); err != nil {
				t.Error(err)
				return
			}
			mu.Lock()
			ordem = append(ordem, p.id)
			mu.Unlock()
			torre.Liberar(p.id)
		})
		for torre.Esperando() != i+1 {
			time.Sleep(time.Millisecond)
		}
	}
	torre.Liberar("A")
	wg.Wait()
	want := []string{"C", "E", "B", "D"} // prioridade 5 primeiro, e FIFO dentro de cada prioridade
	for i := range want {
		if ordem[i] != want[i] {
			t.Fatalf("ordem = %v, quero %v", ordem, want)
		}
	}
}

// 100 aeronaves: NUNCA duas na pista ao mesmo tempo, e todas pousam.
func TestExclusaoMutua(t *testing.T) {
	torre := NovaTorre()
	defer torre.Encerrar()
	var naPista, maximo, pousos atomic.Int32
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			id := string(rune('A'+i%26)) + string(rune('0'+i/26))
			if err := torre.Solicitar(context.Background(), id, i%3); err != nil {
				t.Error(err)
				return
			}
			n := naPista.Add(1)
			for {
				m := maximo.Load()
				if n <= m || maximo.CompareAndSwap(m, n) {
					break
				}
			}
			pousos.Add(1)
			naPista.Add(-1)
			torre.Liberar(id)
		})
	}
	wg.Wait()
	if maximo.Load() != 1 || pousos.Load() != 100 {
		t.Fatalf("maximo na pista = %d, pousos = %d", maximo.Load(), pousos.Load())
	}
}

func TestContextoDaAeronave(t *testing.T) {
	torre := NovaTorre()
	defer torre.Encerrar()
	torre.Solicitar(context.Background(), "A", 1) // ocupa a pista e nao libera
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err := torre.Solicitar(ctx, "B", 1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}
}

func TestTorreEncerrada(t *testing.T) {
	torre := NovaTorre()
	torre.Encerrar()
	if err := torre.Solicitar(context.Background(), "A", 1); !errors.Is(err, ErrTorreEncerrada) {
		t.Fatalf("err = %v", err)
	}
	torre.Liberar("A") // nao pode travar
	if torre.Esperando() != 0 {
		t.Fatal("torre encerrada deveria informar 0")
	}
}
