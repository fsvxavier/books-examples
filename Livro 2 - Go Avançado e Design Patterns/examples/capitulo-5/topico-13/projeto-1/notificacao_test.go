package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// CanalFalso grava o que recebeu e falha as primeiras N vezes.
type CanalFalso struct {
	mu       sync.Mutex
	falhar   int
	Recebido []string
}

func (c *CanalFalso) Nome() string { return "falso" }
func (c *CanalFalso) Enviar(ctx context.Context, destino, texto string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.falhar > 0 {
		c.falhar--
		return ErrTemporario
	}
	c.Recebido = append(c.Recebido, destino+":"+texto)
	return nil
}

func TestSimplesTrocaDeCanal(t *testing.T) {
	a, b := &CanalFalso{}, &CanalFalso{}
	s := NovoSimples(a)
	s.Avisar(context.Background(), "x", "um")
	s.Trocar(b)
	s.Avisar(context.Background(), "x", "dois")
	if len(a.Recebido) != 1 || len(b.Recebido) != 1 {
		t.Fatalf("a=%v b=%v", a.Recebido, b.Recebido)
	}
}

func TestTrocaConcorrente(t *testing.T) {
	a, b := &CanalFalso{}, &CanalFalso{}
	s := NovoSimples(a)
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			if i%10 == 0 {
				s.Trocar(b)
			}
			if err := s.Avisar(context.Background(), "x", "m"); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if total := len(a.Recebido) + len(b.Recebido); total != 100 {
		t.Fatalf("total = %d", total)
	}
}

func TestRetentativaEsperaExponencial(t *testing.T) {
	var esperas []time.Duration
	n := ComRetentativa{
		Canal: &CanalFalso{falhar: 3}, Tentativas: 5, Espera: 100 * time.Millisecond,
		Dormir: func(_ context.Context, d time.Duration) error { esperas = append(esperas, d); return nil },
	}
	if err := n.Avisar(context.Background(), "x", "m"); err != nil {
		t.Fatal(err)
	}
	want := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond}
	if len(esperas) != 3 || esperas[0] != want[0] || esperas[1] != want[1] || esperas[2] != want[2] {
		t.Fatalf("esperas = %v, quero %v", esperas, want)
	}
}

func TestRetentativaDesiste(t *testing.T) {
	n := ComRetentativa{
		Canal: &CanalFalso{falhar: 10}, Tentativas: 3, Espera: time.Millisecond,
		Dormir: func(context.Context, time.Duration) error { return nil },
	}
	err := n.Avisar(context.Background(), "x", "m")
	if !errors.Is(err, ErrTemporario) {
		t.Fatalf("err = %v: a causa deveria continuar na cadeia", err)
	}
}

func TestRetentativaNaoRepeteErroPermanente(t *testing.T) {
	c := &contador{}
	n := ComRetentativa{Canal: c, Tentativas: 5, Espera: time.Millisecond, Dormir: Dormir}
	if err := n.Avisar(context.Background(), "x", "m"); err == nil {
		t.Fatal("esperava erro")
	}
	if c.n != 1 {
		t.Fatalf("tentou %d vezes um erro permanente", c.n)
	}
}

type contador struct{ n int }

func (c *contador) Nome() string { return "contador" }
func (c *contador) Enviar(context.Context, string, string) error {
	c.n++
	return errors.New("destino invalido")
}

func TestRetentativaRespeitaContexto(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	n := ComRetentativa{Canal: &CanalFalso{falhar: 10}, Tentativas: 5, Espera: time.Second, Dormir: Dormir}
	inicio := time.Now()
	err := n.Avisar(ctx, "x", "m")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(inicio) > 500*time.Millisecond {
		t.Fatalf("err = %v apos %v", err, time.Since(inicio))
	}
}

func TestUrgenteSoFalhaSeTodosFalharem(t *testing.T) {
	ok, ruim := &CanalFalso{}, &CanalFalso{falhar: 1}
	var avisos int
	u := Urgente{Canais: []Canal{ok, ruim}, Falhou: func(string, error) { avisos++ }}
	if err := u.Avisar(context.Background(), "x", "m"); err != nil || avisos != 1 {
		t.Fatalf("err=%v avisos=%d", err, avisos)
	}
	u = Urgente{Canais: []Canal{&CanalFalso{falhar: 1}, &CanalFalso{falhar: 1}}}
	if err := u.Avisar(context.Background(), "x", "m"); err == nil {
		t.Fatal("todos falharam e nao houve erro")
	}
}
