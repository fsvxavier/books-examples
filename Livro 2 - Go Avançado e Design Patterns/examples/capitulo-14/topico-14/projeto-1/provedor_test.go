package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"golang.org/x/time/rate"
)

func prov(nome string, lat time.Duration, preco int, cota rate.Limit, burst int) (*Provedor, *atomic.Int32, *atomic.Bool) {
	var n atomic.Int32
	var falha atomic.Bool
	return &Provedor{
		Nome: nome, Timeout: 200 * time.Millisecond,
		Limite: rate.NewLimiter(cota, burst),
		Disj:   &Disjuntor{LimiteFalhas: 3, Espera: time.Minute},
		Preco: func(ctx context.Context, _ string) (int, error) {
			n.Add(1)
			select {
			case <-time.After(lat):
			case <-ctx.Done():
				return 0, ctx.Err()
			}
			if falha.Load() {
				return 0, errors.New("500")
			}
			return preco, nil
		},
	}, &n, &falha
}

func TestEscolheOMelhorEntreOsQueResponderam(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a, _, _ := prov("a", 10*time.Millisecond, 100, 1000, 10)
		b, _, _ := prov("b", 20*time.Millisecond, 80, 1000, 10)
		lento, _, _ := prov("lento", 5*time.Second, 1, 1000, 10) // o mais barato, mas estoura
		c := NovoCotador(time.Minute, 300*time.Millisecond, a, b, lento)
		ini := time.Now()
		q, err := c.Cotar(context.Background(), "x")
		if err != nil || q.Provedor != "b" || q.Preco != 80 {
			t.Fatalf("q=%+v err=%v", q, err)
		}
		if _, ok := q.Falhas["lento"]; !ok {
			t.Fatalf("o lento devia aparecer nas falhas: %v", q.Falhas)
		}
		if d := time.Since(ini); d != 200*time.Millisecond { // o Timeout do provedor
			t.Fatalf("levou %v", d)
		}
	})
}

func TestTodosFalhamEhErro(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a, _, fa := prov("a", time.Millisecond, 1, 1000, 10)
		b, _, fb := prov("b", time.Millisecond, 1, 1000, 10)
		fa.Store(true)
		fb.Store(true)
		c := NovoCotador(time.Minute, time.Second, a, b)
		_, err := c.Cotar(context.Background(), "x")
		if !errors.Is(err, ErrTodosFalharam) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestCacheEColapsoDeCargas(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a, na, _ := prov("a", 10*time.Millisecond, 5, 1000, 10)
		c := NovoCotador(time.Minute, time.Second, a)
		var wg sync.WaitGroup
		for range 200 {
			wg.Go(func() { _, _ = c.Cotar(context.Background(), "x") })
		}
		wg.Wait()
		_, _ = c.Cotar(context.Background(), "x")
		if na.Load() != 1 || c.Cargas() != 1 {
			t.Fatalf("chamadas=%d cargas=%d, quero 1 e 1", na.Load(), c.Cargas())
		}
		time.Sleep(2 * time.Minute) // o TTL vence
		_, _ = c.Cotar(context.Background(), "x")
		if na.Load() != 2 {
			t.Fatalf("depois do TTL: chamadas=%d", na.Load())
		}
	})
}

func TestDisjuntorPoupaOProvedorCaido(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bom, _, _ := prov("bom", time.Millisecond, 10, 1000, 10)
		ruim, nr, fr := prov("ruim", time.Millisecond, 1, 1000, 10)
		fr.Store(true)
		c := NovoCotador(time.Minute, time.Second, bom, ruim)
		for i := range 10 {
			if _, err := c.Cotar(context.Background(), string(rune('a'+i))); err != nil {
				t.Fatal(err)
			}
		}
		if nr.Load() != 3 { // LimiteFalhas = 3: depois disso, nem chama
			t.Fatalf("o provedor caído recebeu %d chamadas, quero 3", nr.Load())
		}
		if ruim.Disj.Estado() != Aberto {
			t.Fatalf("estado %v", ruim.Disj.Estado())
		}
	})
}

func TestLimiteDeTaxaDoProvedor(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a, _, _ := prov("a", time.Millisecond, 5, 10, 1) // 10/s, sem rajada
		c := NovoCotador(time.Minute, 10*time.Second, a)
		ini := time.Now()
		for i := range 6 { // 6 produtos diferentes, sem cache entre eles
			if _, err := c.Cotar(context.Background(), string(rune('a'+i))); err != nil {
				t.Fatal(err)
			}
		}
		// 6 consultas a 10/s: a 1ª é imediata e as outras 5 esperam 100 ms cada.
		if d := time.Since(ini); d < 500*time.Millisecond || d > 520*time.Millisecond {
			t.Fatalf("levou %v, esperava cerca de 500ms (limite de taxa)", d)
		}
	})
}

func TestPrazoCurtoDeQuemEspera(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a, na, _ := prov("a", 150*time.Millisecond, 5, 1000, 10)
		c := NovoCotador(time.Minute, 5*time.Second, a)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, err := c.Cotar(ctx, "x")
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v", err)
		}
		time.Sleep(time.Second) // a carga continuou sem o chamador e guardou o resultado
		q, err := c.Cotar(context.Background(), "x")
		if err != nil || q.Preco != 5 || na.Load() != 1 {
			t.Fatalf("q=%+v err=%v chamadas=%d", q, err, na.Load())
		}
	})
}
