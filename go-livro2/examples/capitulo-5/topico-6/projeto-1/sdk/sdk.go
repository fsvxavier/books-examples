// Package sdk simula o SDK de um banco de terceiros que nao podemos alterar:
// valores em reais (float64), erros proprios, nenhum context.Context.
package sdk

import (
	"fmt"
	"time"
)

type APIError struct {
	Codigo   int
	Mensagem string
}

func (e *APIError) Error() string { return fmt.Sprintf("sdk: erro %d: %s", e.Codigo, e.Mensagem) }

type Cliente struct {
	saldos map[string]float64
	Atraso time.Duration
}

func NovoCliente(saldos map[string]float64) *Cliente {
	return &Cliente{saldos: saldos}
}

func (c *Cliente) ConsultarSaldo(conta string) (float64, error) {
	time.Sleep(c.Atraso)
	s, ok := c.saldos[conta]
	if !ok {
		return 0, &APIError{Codigo: 404, Mensagem: "conta " + conta + " nao existe"}
	}
	return s, nil
}

func (c *Cliente) TransferirReais(origem, destino string, valor float64) (string, error) {
	time.Sleep(c.Atraso)
	so, ok := c.saldos[origem]
	if !ok {
		return "", &APIError{Codigo: 404, Mensagem: "conta " + origem + " nao existe"}
	}
	if _, ok := c.saldos[destino]; !ok {
		return "", &APIError{Codigo: 404, Mensagem: "conta " + destino + " nao existe"}
	}
	if so < valor {
		return "", &APIError{Codigo: 422, Mensagem: "saldo da origem menor que o valor"}
	}
	c.saldos[origem] -= valor
	c.saldos[destino] += valor
	return fmt.Sprintf("OP-%s-%s", origem, destino), nil
}
