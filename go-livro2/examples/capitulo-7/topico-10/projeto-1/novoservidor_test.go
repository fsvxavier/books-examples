package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var (
	ana  = Perfil{1, "Ana", "ana@x.com"}
	bia  = Perfil{2, "Bia", "bia@x.com"}
	ctxA = ComUsuario(context.Background(), Usuario{ID: 1})
	ctxB = ComUsuario(context.Background(), Usuario{ID: 2})
)

// TesteDeContrato vale para QUALQUER implementacao de Perfis: banco, proxies e cliente remoto.
func TesteDeContrato(t *testing.T, nova func(t *testing.T) Perfis) {
	t.Helper()
	t.Run("obtem o proprio perfil", func(t *testing.T) {
		got, err := nova(t).Obter(ctxA, 1)
		if err != nil || got != ana {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	})
	t.Run("perfil inexistente", func(t *testing.T) {
		_, err := nova(t).Obter(ComUsuario(context.Background(), Usuario{ID: 9, Admin: true}), 99)
		if !errors.Is(err, ErrNaoEncontrado) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("atualiza e le de volta", func(t *testing.T) {
		s := nova(t)
		if err := s.Atualizar(ctxA, Perfil{1, "Ana Maria", "ana@x.com"}); err != nil {
			t.Fatal(err)
		}
		got, _ := s.Obter(ctxA, 1)
		if got.Nome != "Ana Maria" {
			t.Fatalf("got = %+v", got)
		}
	})
}

func novoBanco(t *testing.T) Perfis { return NovoBanco(ana, bia) }

func TestContratoBanco(t *testing.T) { TesteDeContrato(t, novoBanco) }

func TestContratoRemoto(t *testing.T) {
	TesteDeContrato(t, func(t *testing.T) Perfis {
		srv := httptest.NewServer(NovoServidor(NovoBanco(ana, bia)))
		t.Cleanup(srv.Close)
		return NovoClienteRemoto(srv.URL, srv.Client())
	})
}

func TestContratoPilhaCompleta(t *testing.T) {
	TesteDeContrato(t, func(t *testing.T) Perfis {
		return Proteger(ComCache(ComLog("x", NovoBanco(ana, bia)), time.Minute, time.Now))
	})
}

func TestProtecaoViaRemoto(t *testing.T) {
	// servidor protegido; o cliente remoto propaga a identidade.
	srv := httptest.NewServer(NovoServidor(Proteger(NovoBanco(ana, bia))))
	defer srv.Close()
	c := NovoClienteRemoto(srv.URL, srv.Client())
	if _, err := c.Obter(ctxA, 2); !errors.Is(err, ErrProibido) {
		t.Fatalf("err = %v", err)
	}
	if _, err := c.Obter(ctxB, 2); err != nil {
		t.Fatalf("o proprio dono deveria ler: %v", err)
	}
}

func TestPrazoNoRemoto(t *testing.T) {
	b := NovoBanco(ana)
	b.Latencia = 300 * time.Millisecond
	srv := httptest.NewServer(NovoServidor(b))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(ctxA, 30*time.Millisecond)
	defer cancel()
	inicio := time.Now()
	_, err := NovoClienteRemoto(srv.URL, srv.Client()).Obter(ctx, 1)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(inicio) > 200*time.Millisecond {
		t.Fatalf("err=%v apos %v", err, time.Since(inicio))
	}
}

func TestCacheNaoGuardaErro(t *testing.T) {
	b := NovoBanco(ana)
	c := ComCache(b, time.Minute, time.Now)
	for range 3 {
		if _, err := c.Obter(ctxA, 99); !errors.Is(err, ErrNaoEncontrado) {
			t.Fatalf("err = %v", err)
		}
	}
	if b.Leituras.Load() != 3 {
		t.Fatalf("leituras = %d: o erro foi guardado no cache", b.Leituras.Load())
	}
}

func TestCacheConcorrente(t *testing.T) {
	b := NovoBanco(ana, bia)
	c := ComCache(b, time.Minute, time.Now)
	var wg sync.WaitGroup
	for i := range 200 {
		wg.Go(func() {
			id := 1 + i%2
			if _, err := c.Obter(ctxA, id); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if b.Leituras.Load() > 2 {
		t.Fatalf("leituras = %d, esperava no maximo 2 (uma por id)", b.Leituras.Load())
	}
}
