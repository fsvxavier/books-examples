package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type relogio struct {
	mu sync.Mutex
	t  time.Time
}

func (r *relogio) Agora() time.Time { r.mu.Lock(); defer r.mu.Unlock(); return r.t }
func (r *relogio) Avancar(d time.Duration) {
	r.mu.Lock()
	r.t = r.t.Add(d)
	r.mu.Unlock()
}

func novoCache(r *relogio, carregar func(context.Context, string) (int, error)) *Cache[int] {
	return &Cache[int]{TTL: time.Minute, Agora: r.Agora, Carregar: carregar, ProtegerCarga: true, ServirVelho: true}
}

func TestTTLExpira(t *testing.T) {
	r := &relogio{t: time.Unix(0, 0)}
	n := 0
	c := novoCache(r, func(context.Context, string) (int, error) { n++; return n, nil })
	v1, _, _ := c.Obter(context.Background(), "a")
	v2, _, _ := c.Obter(context.Background(), "a")
	r.Avancar(61 * time.Second)
	v3, _, _ := c.Obter(context.Background(), "a")
	if v1 != 1 || v2 != 1 || v3 != 2 {
		t.Fatalf("valores %d %d %d", v1, v2, v3)
	}
}

func TestErroNaoEhGuardado(t *testing.T) {
	r := &relogio{t: time.Unix(0, 0)}
	falha := true
	c := novoCache(r, func(context.Context, string) (int, error) {
		if falha {
			return 0, errors.New("fora do ar")
		}
		return 7, nil
	})
	if _, _, err := c.Obter(context.Background(), "a"); err == nil {
		t.Fatal("esperava erro")
	}
	falha = false
	v, _, err := c.Obter(context.Background(), "a")
	if err != nil || v != 7 {
		t.Fatalf("depois do erro: %d %v (o erro foi guardado?)", v, err)
	}
}

func TestServeVelhoQuandoRecargaFalha(t *testing.T) {
	r := &relogio{t: time.Unix(0, 0)}
	falha := false
	c := novoCache(r, func(context.Context, string) (int, error) {
		if falha {
			return 0, errors.New("fora do ar")
		}
		return 7, nil
	})
	_, _, _ = c.Obter(context.Background(), "a")
	r.Avancar(2 * time.Minute)
	falha = true
	v, velho, err := c.Obter(context.Background(), "a")
	if err != nil || v != 7 || !velho {
		t.Fatalf("v=%d velho=%v err=%v", v, velho, err)
	}
}

// Determinístico: o carregador só termina quando os 50 chamadores já
// estão esperando; com a proteção, deve haver UMA carga.
func TestUmaCargaParaVariosChamadores(t *testing.T) {
	r := &relogio{t: time.Unix(0, 0)}
	liberar := make(chan struct{})
	var cargas atomic.Int32
	c := novoCache(r, func(context.Context, string) (int, error) {
		cargas.Add(1)
		<-liberar
		return 1, nil
	})
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() { _, _, _ = c.Obter(context.Background(), "k") })
	}
	time.Sleep(50 * time.Millisecond) // todos entraram e esperam
	close(liberar)
	wg.Wait()
	if cargas.Load() != 1 {
		t.Fatalf("cargas = %d, quero 1", cargas.Load())
	}
}

func TestChavesDiferentesNaoSeBloqueiam(t *testing.T) {
	r := &relogio{t: time.Unix(0, 0)}
	liberarA := make(chan struct{})
	c := novoCache(r, func(_ context.Context, k string) (int, error) {
		if k == "a" {
			<-liberarA
		}
		return 1, nil
	})
	go func() { _, _, _ = c.Obter(context.Background(), "a") }()
	time.Sleep(10 * time.Millisecond)
	feito := make(chan struct{})
	go func() { _, _, _ = c.Obter(context.Background(), "b"); close(feito) }()
	select {
	case <-feito:
	case <-time.After(time.Second):
		t.Fatal("a chave b ficou presa esperando a chave a")
	}
	close(liberarA)
}

func TestChamadorDesisteSemMatarACarga(t *testing.T) {
	r := &relogio{t: time.Unix(0, 0)}
	liberar := make(chan struct{})
	c := novoCache(r, func(context.Context, string) (int, error) { <-liberar; return 9, nil })

	ctxCurto, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, _, err := c.Obter(ctxCurto, "k"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}
	close(liberar)
	time.Sleep(20 * time.Millisecond) // a carga terminou sozinha e guardou
	v, _, err := c.Obter(context.Background(), "k")
	if err != nil || v != 9 || c.Cargas.Load() != 1 {
		t.Fatalf("v=%d err=%v cargas=%d", v, err, c.Cargas.Load())
	}
}
