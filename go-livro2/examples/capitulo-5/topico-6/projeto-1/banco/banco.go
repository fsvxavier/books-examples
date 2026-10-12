// Package banco define o que o NOSSO sistema espera de um banco e adapta o SDK.
package banco

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-5/topico-6/projeto-1/sdk"
)

var (
	ErrContaInexistente  = errors.New("conta inexistente")
	ErrSaldoInsuficiente = errors.New("saldo insuficiente")
)

// Banco e a interface do cliente: dinheiro em centavos (int64), erros nossos, context.
type Banco interface {
	Saldo(ctx context.Context, conta string) (centavos int64, err error)
	Transferir(ctx context.Context, origem, destino string, centavos int64) (operacao string, err error)
}

// AdaptadorSDK traduz o Banco para o SDK de terceiros.
type AdaptadorSDK struct {
	cli *sdk.Cliente
}

func NovoAdaptadorSDK(cli *sdk.Cliente) *AdaptadorSDK { return &AdaptadorSDK{cli: cli} }

func paraCentavos(reais float64) int64 { return int64(math.Round(reais * 100)) }
func paraReais(centavos int64) float64 { return float64(centavos) / 100 }

// traduzir converte o erro do SDK em erro do nosso dominio, preservando a causa.
func traduzir(err error) error {
	var api *sdk.APIError
	if errors.As(err, &api) {
		switch api.Codigo {
		case 404:
			return fmt.Errorf("%w: %s", ErrContaInexistente, api.Mensagem)
		case 422:
			return fmt.Errorf("%w: %s", ErrSaldoInsuficiente, api.Mensagem)
		}
	}
	return err
}

type resultado[T any] struct {
	valor T
	err   error
}

// comContexto roda uma chamada bloqueante do SDK sem context e a abandona se ctx acabar.
func comContexto[T any](ctx context.Context, f func() (T, error)) (T, error) {
	ch := make(chan resultado[T], 1) // buffer 1: a goroutine nunca fica presa
	go func() {
		v, err := f()
		ch <- resultado[T]{v, err}
	}()
	select {
	case r := <-ch:
		return r.valor, r.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

func (a *AdaptadorSDK) Saldo(ctx context.Context, conta string) (int64, error) {
	reais, err := comContexto(ctx, func() (float64, error) { return a.cli.ConsultarSaldo(conta) })
	if err != nil {
		return 0, traduzir(err)
	}
	return paraCentavos(reais), nil
}

func (a *AdaptadorSDK) Transferir(ctx context.Context, origem, destino string, centavos int64) (string, error) {
	op, err := comContexto(ctx, func() (string, error) {
		return a.cli.TransferirReais(origem, destino, paraReais(centavos))
	})
	if err != nil {
		return "", traduzir(err)
	}
	return op, nil
}
