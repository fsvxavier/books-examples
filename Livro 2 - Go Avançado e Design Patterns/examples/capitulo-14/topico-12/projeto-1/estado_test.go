package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

var errX = errors.New("falha")

func falha(context.Context) error { return errX }
func ok(context.Context) error    { return nil }

func TestAbreDepoisDeFalhasSeguidas(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d := &Disjuntor{LimiteFalhas: 3, Espera: time.Minute}
		for range 3 {
			_ = d.Executar(context.Background(), falha)
		}
		if d.Estado() != Aberto {
			t.Fatalf("estado = %v", d.Estado())
		}
		chamou := false
		err := d.Executar(context.Background(), func(context.Context) error { chamou = true; return nil })
		if !errors.Is(err, ErrAberto) || chamou {
			t.Fatalf("err=%v chamou=%v: aberto deve recusar sem chamar", err, chamou)
		}
	})
}

func TestSucessoZeraAContagem(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d := &Disjuntor{LimiteFalhas: 3, Espera: time.Minute}
		_ = d.Executar(context.Background(), falha)
		_ = d.Executar(context.Background(), falha)
		_ = d.Executar(context.Background(), ok) // quebra a sequência
		_ = d.Executar(context.Background(), falha)
		_ = d.Executar(context.Background(), falha)
		if d.Estado() != Fechado {
			t.Fatalf("estado = %v: só falhas SEGUIDAS devem abrir", d.Estado())
		}
	})
}

func TestMeioAbertoDepoisDaEspera(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d := &Disjuntor{LimiteFalhas: 1, Espera: time.Minute}
		_ = d.Executar(context.Background(), falha)
		time.Sleep(59 * time.Second)
		if err := d.Executar(context.Background(), ok); !errors.Is(err, ErrAberto) {
			t.Fatalf("antes da espera: %v", err)
		}
		time.Sleep(2 * time.Second)
		if err := d.Executar(context.Background(), ok); err != nil {
			t.Fatalf("a sonda devia passar: %v", err)
		}
		if d.Estado() != Fechado {
			t.Fatalf("sonda bem-sucedida deve fechar: %v", d.Estado())
		}
	})
}

func TestSondaQueFalhaReabre(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d := &Disjuntor{LimiteFalhas: 1, Espera: time.Minute}
		_ = d.Executar(context.Background(), falha)
		time.Sleep(time.Minute)
		_ = d.Executar(context.Background(), falha) // a sonda falha
		if d.Estado() != Aberto {
			t.Fatalf("estado = %v", d.Estado())
		}
		time.Sleep(30 * time.Second) // a espera recomeçou do zero
		if err := d.Executar(context.Background(), ok); !errors.Is(err, ErrAberto) {
			t.Fatalf("a espera deve recomeçar: %v", err)
		}
	})
}

func TestSoUmaSonda(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d := &Disjuntor{LimiteFalhas: 1, Espera: time.Minute}
		_ = d.Executar(context.Background(), falha)
		time.Sleep(time.Minute)

		liberar := make(chan struct{})
		var wg sync.WaitGroup
		var passaram, recusadas int
		var mu sync.Mutex
		for range 10 {
			wg.Go(func() {
				err := d.Executar(context.Background(), func(context.Context) error { <-liberar; return nil })
				mu.Lock()
				if errors.Is(err, ErrAberto) {
					recusadas++
				} else {
					passaram++
				}
				mu.Unlock()
			})
		}
		synctest.Wait()
		close(liberar)
		wg.Wait()
		if passaram != 1 || recusadas != 9 {
			t.Fatalf("passaram=%d recusadas=%d: só UMA sonda por vez", passaram, recusadas)
		}
	})
}

func TestErroQueNaoContaNaoAbre(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d := &Disjuntor{LimiteFalhas: 2, Espera: time.Minute,
			Conta: func(err error) bool { return !errors.Is(err, context.Canceled) }}
		for range 10 {
			_ = d.Executar(context.Background(), func(context.Context) error { return context.Canceled })
		}
		if d.Estado() != Fechado {
			t.Fatalf("o cancelamento do chamador não é falha da dependência: %v", d.Estado())
		}
	})
}

func TestConcorrenciaComRace(t *testing.T) {
	d := &Disjuntor{LimiteFalhas: 5, Espera: time.Millisecond}
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Go(func() {
			for i := range 500 {
				_ = d.Executar(context.Background(), func(context.Context) error {
					if (i+g)%3 == 0 {
						return errX
					}
					return nil
				})
			}
		})
	}
	wg.Wait()
}
