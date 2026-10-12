package banco

import (
	"context"
	"errors"
	"testing"
	"time"

	"exemplos/livro2/capitulo-5/topico-6/projeto-1/sdk"
)

func novo() (*AdaptadorSDK, *sdk.Cliente) {
	cli := sdk.NovoCliente(map[string]float64{"ana": 100.10, "bia": 0.20})
	return NovoAdaptadorSDK(cli), cli
}

func TestConversaoDeUnidades(t *testing.T) {
	a, _ := novo()
	got, err := a.Saldo(context.Background(), "ana")
	if err != nil || got != 10010 {
		t.Fatalf("Saldo = %d, %v; quero 10010", got, err)
	}
}

// Truncar em vez de arredondar erraria 1 centavo (com variaveis, 19.99*100 = 1998.9999999999998).
func TestArredondamentoNaoTrunca(t *testing.T) {
	reais := 19.99 // variavel: com constantes o compilador calcula com precisao exata
	if got := int64(reais * 100); got != 1998 {
		t.Fatalf("a premissa do teste mudou: truncar deu %d", got)
	}
	if got := paraCentavos(reais); got != 1999 {
		t.Fatalf("paraCentavos(19.99) = %d, quero 1999", got)
	}
}

func TestTransferir(t *testing.T) {
	a, _ := novo()
	op, err := a.Transferir(context.Background(), "ana", "bia", 1000)
	if err != nil || op != "OP-ana-bia" {
		t.Fatalf("op=%q err=%v", op, err)
	}
	if s, _ := a.Saldo(context.Background(), "bia"); s != 1020 {
		t.Fatalf("saldo de bia = %d", s)
	}
}

func TestErrosTraduzidos(t *testing.T) {
	a, _ := novo()
	ctx := context.Background()
	_, err := a.Saldo(ctx, "zeca")
	if !errors.Is(err, ErrContaInexistente) {
		t.Fatalf("err = %v", err)
	}
	var api *sdk.APIError
	if errors.As(err, &api) {
		t.Fatalf("o erro do SDK vazou (a causa nao deveria estar na cadeia): %v", err)
	}
	_, err = a.Transferir(ctx, "bia", "ana", 999999)
	if !errors.Is(err, ErrSaldoInsuficiente) {
		t.Fatalf("err = %v", err)
	}
}

func TestContextoCancela(t *testing.T) {
	a, cli := novo()
	cli.Atraso = 300 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	inicio := time.Now()
	_, err := a.Saldo(ctx, "ana")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}
	if d := time.Since(inicio); d > 200*time.Millisecond {
		t.Fatalf("demorou %v: o adaptador esperou o SDK", d)
	}
}
