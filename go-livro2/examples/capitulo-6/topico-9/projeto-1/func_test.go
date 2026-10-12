package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestMemoizarNaoGuardaErro(t *testing.T) {
	var n int
	f := Aplicar(func(_ context.Context, a int) (int, error) {
		n++
		if n == 1 {
			return 0, errors.New("falha")
		}
		return a * 2, nil
	}, Memoizar[int, int]())
	if _, err := f(context.Background(), 5); err == nil {
		t.Fatal("esperava erro na 1a")
	}
	if v, err := f(context.Background(), 5); err != nil || v != 10 {
		t.Fatalf("2a: %v %v", v, err)
	}
	f(context.Background(), 5)
	if n != 2 {
		t.Fatalf("chamadas = %d, quero 2 (a 3a veio do cache)", n)
	}
}

func TestRetentarPreservaCausa(t *testing.T) {
	sentinela := errors.New("boom")
	f := Aplicar(func(context.Context, int) (int, error) { return 0, sentinela }, Retentar[int, int](3))
	if err := func() error { _, e := f(context.Background(), 1); return e }(); !errors.Is(err, sentinela) {
		t.Fatalf("err = %v", err)
	}
}

func TestRetentarRespeitaContexto(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var n int
	f := Aplicar(func(ctx context.Context, _ int) (int, error) {
		n++
		cancel()
		return 0, errors.New("x")
	}, Retentar[int, int](10))
	f(ctx, 1)
	if n != 1 {
		t.Fatalf("tentou %d vezes depois do cancelamento", n)
	}
}

func TestMemoizarConcorrente(t *testing.T) {
	var n atomic.Int64
	f := Aplicar(func(_ context.Context, a int) (int, error) { n.Add(1); return a, nil }, Memoizar[int, int]())
	var wg sync.WaitGroup
	for i := range 200 {
		wg.Go(func() {
			if v, _ := f(context.Background(), i%5); v != i%5 {
				t.Errorf("v = %d", v)
			}
		})
	}
	wg.Wait()
	t.Logf("chamadas a funcao cara com 200 concorrentes e 5 chaves: %d (duplicadas sao possiveis)", n.Load())
	if n.Load() < 5 {
		t.Fatalf("chamadas = %d", n.Load())
	}
}
