package main

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestOrdemPreservada(t *testing.T) {
	itens := make([]int, 200)
	for i := range itens {
		itens[i] = i
	}
	r, err := Mapear(context.Background(), itens, 8, func(_ context.Context, i, n int) (int, error) {
		time.Sleep(time.Duration((200-n)%7) * 100 * time.Microsecond) // tempos desencontrados
		return n * 2, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range r {
		if v != i*2 {
			t.Fatalf("posição %d = %d", i, v)
		}
	}
}

func TestRespeitaLimite(t *testing.T) {
	for _, limite := range []int{1, 2, 5} {
		var ativas, maxAtivas atomic.Int32
		_, err := Mapear(context.Background(), make([]int, 40), limite, func(_ context.Context, _, _ int) (int, error) {
			n := ativas.Add(1)
			for {
				m := maxAtivas.Load()
				if n <= m || maxAtivas.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			ativas.Add(-1)
			return 0, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if int(maxAtivas.Load()) > limite {
			t.Fatalf("limite %d, máximo simultâneo %d", limite, maxAtivas.Load())
		}
	}
}

func TestPrimeiroErroPoupaTarefas(t *testing.T) {
	var iniciadas atomic.Int32
	_, err := Mapear(context.Background(), make([]int, 1000), 4, func(ctx context.Context, i, _ int) (int, error) {
		iniciadas.Add(1)
		if i == 5 {
			return 0, ErroDeNegocio
		}
		time.Sleep(time.Millisecond)
		return 0, nil
	})
	if !errors.Is(err, ErroDeNegocio) {
		t.Fatalf("err = %v", err)
	}
	if iniciadas.Load() > 50 {
		t.Fatalf("iniciou %d tarefas depois do erro", iniciadas.Load())
	}
}

func TestPanicViraErro(t *testing.T) {
	_, err := Mapear(context.Background(), []int{1, 2, 3}, 0, func(_ context.Context, i, _ int) (int, error) {
		if i == 1 {
			panic("boom")
		}
		return 0, nil
	})
	var ep *ErroPanic
	if !errors.As(err, &ep) || ep.Valor != "boom" || len(ep.Pilha) == 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestContextoCanceladoDeFora(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var iniciadas atomic.Int32
	r, err := Mapear(ctx, []int{1, 2, 3}, 2, func(_ context.Context, _, n int) (int, error) {
		iniciadas.Add(1)
		return n, nil
	})
	if !errors.Is(err, context.Canceled) || r != nil {
		t.Fatalf("r=%v err=%v: cancelado antes de começar deve ser erro, não sucesso vazio", r, err)
	}
	if iniciadas.Load() != 0 {
		t.Fatalf("%d tarefas iniciadas com o contexto já cancelado", iniciadas.Load())
	}
}

func TestSemVazamentoDeGoroutines(t *testing.T) {
	antes := runtime.NumGoroutine()
	for range 20 {
		_, _ = Mapear(context.Background(), make([]int, 50), 4, func(ctx context.Context, i, _ int) (int, error) {
			if i == 7 {
				return 0, ErroDeNegocio
			}
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(time.Millisecond):
				return 0, nil
			}
		})
	}
	time.Sleep(20 * time.Millisecond)
	if depois := runtime.NumGoroutine(); depois > antes {
		t.Fatalf("goroutines: %d antes, %d depois", antes, depois)
	}
}

func TestVazio(t *testing.T) {
	r, err := Mapear(context.Background(), []int(nil), 3, func(context.Context, int, int) (int, error) { return 0, nil })
	if err != nil || len(r) != 0 {
		t.Fatalf("%v %v", r, err)
	}
}
