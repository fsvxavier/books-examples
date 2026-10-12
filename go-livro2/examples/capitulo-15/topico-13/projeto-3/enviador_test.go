package bom

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

type enviadorFalso struct {
	chamadas   atomic.Int32
	falhasAnte int32 // quantas chamadas iniciais devem falhar
	atraso     time.Duration
}

func (e *enviadorFalso) Enviar(ctx context.Context, msg string) error {
	n := e.chamadas.Add(1)
	time.Sleep(e.atraso)
	if n <= e.falhasAnte {
		return errors.New("smtp indisponível")
	}
	return nil
}

var opcoes = Opcoes{Capacidade: 4, Tentativas: 3, Espera: 100 * time.Millisecond}

func TestEnviaTudoEEncerra(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		env := &enviadorFalso{atraso: 5 * time.Millisecond}
		n := Novo(env, opcoes)
		ctx, cancel := context.WithCancel(t.Context())
		n.Iniciar(ctx)
		for range 4 {
			if err := n.Enfileirar("olá"); err != nil {
				t.Fatal(err)
			}
		}
		time.Sleep(time.Second)
		cancel()
		n.Esperar() // se sobrar goroutine, synctest.Test acusa deadlock
		if n.Enviados() != 4 {
			t.Fatalf("enviados = %d", n.Enviados())
		}
	})
}

func TestTentativasComEspera(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		env := &enviadorFalso{falhasAnte: 2}
		n := Novo(env, opcoes)
		ctx, cancel := context.WithCancel(t.Context())
		n.Iniciar(ctx)
		inicio := time.Now()
		n.Enfileirar("olá")
		time.Sleep(time.Second)
		cancel()
		n.Esperar()
		t.Logf("chamadas=%d enviados=%d (tempo falso: %v)", env.chamadas.Load(), n.Enviados(), time.Since(inicio))
		if env.chamadas.Load() != 3 || n.Enviados() != 1 {
			t.Fatal("deveria ter tentado 3 vezes e enviado 1")
		}
	})
}

func TestFalhaDefinitivaViraErro(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		env := &enviadorFalso{falhasAnte: 99}
		n := Novo(env, opcoes)
		ctx, cancel := context.WithCancel(t.Context())
		n.Iniciar(ctx)
		n.Enfileirar("olá")
		time.Sleep(time.Second)
		cancel()
		n.Esperar()
		if len(n.Falhas()) != 1 {
			t.Fatalf("falhas = %v", n.Falhas())
		}
		t.Log(n.Falhas()[0])
	})
}

func TestFilaCheiaEMensagemVazia(t *testing.T) {
	n := Novo(&enviadorFalso{}, opcoes) // sem Iniciar: nada consome
	if err := n.Enfileirar(""); !errors.Is(err, ErrMensagemVazia) {
		t.Fatalf("err = %v", err)
	}
	for range opcoes.Capacidade {
		n.Enfileirar("x")
	}
	if err := n.Enfileirar("x"); !errors.Is(err, ErrFilaCheia) {
		t.Fatalf("err = %v", err)
	}
}

func TestEncerraEsvaziandoAFila(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		env := &enviadorFalso{atraso: 50 * time.Millisecond}
		n := Novo(env, opcoes)
		ctx, cancel := context.WithCancel(t.Context())
		n.Iniciar(ctx)
		for range 4 {
			n.Enfileirar("olá")
		}
		time.Sleep(10 * time.Millisecond) // o trabalhador pegou a primeira
		cancel()                          // pedido de parada com mensagens ainda na fila
		n.Esperar()
		if n.Enviados() != 4 {
			t.Fatalf("enviados = %d, a fila deveria ter sido esvaziada", n.Enviados())
		}
	})
}
