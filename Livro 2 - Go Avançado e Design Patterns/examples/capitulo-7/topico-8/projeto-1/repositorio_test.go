package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNaoAbreAteSerUsado(t *testing.T) {
	var n atomic.Int32
	p := NovoProxyPreguicoso(func(context.Context) (Repositorio, error) { n.Add(1); return &repoReal{}, nil })
	if n.Load() != 0 {
		t.Fatal("abriu na construcao")
	}
	p.Buscar(context.Background(), 1)
	p.Buscar(context.Background(), 2)
	if n.Load() != 1 {
		t.Fatalf("aberturas = %d", n.Load())
	}
}

func TestConcorrenteAbreUmaVez(t *testing.T) {
	var n atomic.Int32
	p := NovoProxyPreguicoso(func(context.Context) (Repositorio, error) {
		n.Add(1)
		time.Sleep(20 * time.Millisecond)
		return &repoReal{}, nil
	})
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			if _, err := p.Buscar(context.Background(), i); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if n.Load() != 1 {
		t.Fatalf("aberturas = %d, quero 1", n.Load())
	}
}

func TestErroDeAberturaNaoFicaGuardado(t *testing.T) {
	var n int
	p := NovoProxyPreguicoso(func(context.Context) (Repositorio, error) {
		n++
		if n == 1 {
			return nil, ErrIndisponivel
		}
		return &repoReal{}, nil
	})
	if _, err := p.Buscar(context.Background(), 1); !errors.Is(err, ErrIndisponivel) {
		t.Fatalf("err = %v", err)
	}
	if _, err := p.Buscar(context.Background(), 1); err != nil {
		t.Fatalf("2a tentativa: %v", err)
	}
}

type fechavel struct{ fechou *atomic.Bool }

func (f fechavel) Buscar(context.Context, int) (string, error) { return "", nil }
func (f fechavel) Close() error                                { f.fechou.Store(true); return nil }

func TestCloseSoSeAbriu(t *testing.T) {
	fechou := &atomic.Bool{}
	p := NovoProxyPreguicoso(func(context.Context) (Repositorio, error) { return fechavel{fechou}, nil })
	p.Close()
	if fechou.Load() {
		t.Fatal("fechou algo que nunca abriu")
	}
	p.Buscar(context.Background(), 1)
	p.Close()
	if !fechou.Load() {
		t.Fatal("nao fechou o que abriu")
	}
}

func TestContextoCancelaAAbertura(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	p := NovoProxyPreguicoso(abrir)
	_, err := p.Buscar(ctx, 1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}
}
