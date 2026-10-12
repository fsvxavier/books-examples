package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

type gravador struct {
	mu      sync.Mutex
	lotes   [][]int
	falhar  error
	entrada time.Time
}

func (g *gravador) Gravar(_ context.Context, it []int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.lotes = append(g.lotes, append([]int(nil), it...))
	return g.falhar
}

// O tempo dentro da bolha é falso: 50 ms de espera passam instantaneamente.
func TestGravaQuandoJuntaMaxItens(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g := &gravador{}
		l := &Lote[int]{MaxItens: 3, MaxEspera: time.Hour, Gravar: g.Gravar}
		var wg sync.WaitGroup
		for i := range 3 {
			wg.Go(func() { _ = l.Enviar(context.Background(), i) })
		}
		wg.Wait()
		if len(g.lotes) != 1 || len(g.lotes[0]) != 3 {
			t.Fatalf("lotes = %v", g.lotes)
		}
		l.Fechar()
	})
}

func TestGravaQuandoPassaMaxEspera(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g := &gravador{}
		l := &Lote[int]{MaxItens: 100, MaxEspera: 50 * time.Millisecond, Gravar: g.Gravar}
		ini := time.Now()
		_ = l.Enviar(context.Background(), 1) // sozinho: espera o prazo
		if d := time.Since(ini); d != 50*time.Millisecond {
			t.Fatalf("esperou %v, quero exatamente 50ms", d)
		}
		if len(g.lotes) != 1 {
			t.Fatalf("lotes = %v", g.lotes)
		}
		l.Fechar()
	})
}

func TestFecharGravaOPendente(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g := &gravador{}
		l := &Lote[int]{MaxItens: 100, MaxEspera: time.Hour, Gravar: g.Gravar}
		erro := make(chan error, 1)
		go func() { erro <- l.Enviar(context.Background(), 7) }()
		synctest.Wait() // o item já entrou e está esperando
		l.Fechar()
		if err := <-erro; err != nil {
			t.Fatal(err)
		}
		if len(g.lotes) != 1 || g.lotes[0][0] != 7 {
			t.Fatalf("lotes = %v", g.lotes)
		}
		if err := l.Enviar(context.Background(), 8); !errors.Is(err, ErrFechado) {
			t.Fatalf("depois de fechar: %v", err)
		}
	})
}

func TestErroDoLoteChegaATodos(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		falha := errors.New("banco caiu")
		g := &gravador{falhar: falha}
		l := &Lote[int]{MaxItens: 4, MaxEspera: time.Hour, Gravar: g.Gravar}
		var wg sync.WaitGroup
		erros := make([]error, 4)
		for i := range 4 {
			wg.Go(func() { erros[i] = l.Enviar(context.Background(), i) })
		}
		wg.Wait()
		for i, e := range erros {
			if !errors.Is(e, falha) {
				t.Fatalf("chamador %d: %v", i, e)
			}
		}
		l.Fechar()
	})
}

func TestPanicNaGravacaoViraErro(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := &Lote[int]{MaxItens: 1, MaxEspera: time.Hour, Gravar: func(context.Context, []int) error { panic("boom") }}
		if err := l.Enviar(context.Background(), 1); err == nil {
			t.Fatal("esperava erro")
		}
		l.Fechar()
	})
}

// Desistir não retira o item do lote: ele pode ser gravado mesmo assim.
func TestDesistirNaoRetiraOItem(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g := &gravador{}
		l := &Lote[int]{MaxItens: 100, MaxEspera: 100 * time.Millisecond, Gravar: g.Gravar}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		err := l.Enviar(ctx, 42)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v", err)
		}
		l.Fechar()
		if len(g.lotes) != 1 || g.lotes[0][0] != 42 {
			t.Fatalf("o item 42 devia ter sido gravado mesmo assim: %v", g.lotes)
		}
	})
}
