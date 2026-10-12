package frete

import (
	"context"
	"errors"
	"testing"
	"time"
)

var pedido = Pedido{CEPOrigem: "01310100", CEPDestino: "20040020", PesoGramas: 2500}

// TesteDeContrato vale para QUALQUER implementacao de Transportadora:
// e ele que garante que os adaptadores sao intercambiaveis.
func TesteDeContrato(t *testing.T, nova func() Transportadora) {
	t.Helper()
	t.Run("nome nao vazio", func(t *testing.T) {
		if nova().Nome() == "" {
			t.Fatal("Nome() vazio")
		}
	})
	t.Run("oferta valida", func(t *testing.T) {
		o, err := nova().Cotar(context.Background(), pedido)
		if err != nil {
			t.Fatal(err)
		}
		if o.Centavos <= 0 || o.Prazo <= 0 || o.Transportadora != nova().Nome() {
			t.Fatalf("oferta invalida: %+v", o)
		}
	})
	t.Run("respeita contexto ja cancelado", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := nova().Cotar(ctx, pedido)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, quero context.Canceled", err)
		}
	})
}

func TestContratoCorreios(t *testing.T) {
	TesteDeContrato(t, func() Transportadora { return AdaptadorCorreios{} })
}

func TestContratoRapidoLog(t *testing.T) {
	TesteDeContrato(t, func() Transportadora { return NovoRapidoLog(5 * time.Millisecond) })
}

func TestContratoMoto(t *testing.T) {
	TesteDeContrato(t, func() Transportadora { return NovaMoto(false) })
}

func TestConversoes(t *testing.T) {
	o, err := AdaptadorCorreios{}.Cotar(context.Background(), pedido)
	if err != nil || o.Centavos != 1750 || o.Prazo != 72*time.Hour {
		t.Fatalf("correios: %+v %v", o, err)
	}
	o, err = NovoRapidoLog(0).Cotar(context.Background(), pedido)
	if err != nil || o.Centavos != 2065 || o.Prazo != 48*time.Hour {
		t.Fatalf("rapido: %+v %v", o, err)
	}
	o, err = NovaMoto(false).Cotar(context.Background(), pedido)
	if err != nil || o.Centavos != 2375 || o.Prazo != 24*time.Hour {
		t.Fatalf("moto: %+v %v", o, err)
	}
}

func TestCotarOrdenaPorPrecoEJuntaErros(t *testing.T) {
	ts := []Transportadora{NovaMoto(false), NovaMoto(true), AdaptadorCorreios{}, NovoRapidoLog(0)}
	ofertas, err := Cotar(context.Background(), ts, pedido)
	if len(ofertas) != 3 {
		t.Fatalf("ofertas = %d", len(ofertas))
	}
	for i := 1; i < len(ofertas); i++ {
		if ofertas[i-1].Centavos > ofertas[i].Centavos {
			t.Fatalf("fora de ordem: %+v", ofertas)
		}
	}
	if err == nil || err.Error() != "moto-express: servico fora do ar" {
		t.Fatalf("err = %v", err)
	}
}

func TestCotarComPrazo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	inicio := time.Now()
	ofertas, err := Cotar(ctx, []Transportadora{AdaptadorCorreios{}, NovoRapidoLog(time.Second)}, pedido)
	if time.Since(inicio) > 500*time.Millisecond {
		t.Fatal("esperou a transportadora lenta")
	}
	if len(ofertas) != 1 || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ofertas=%d err=%v", len(ofertas), err)
	}
}
