package main

import (
	"context"
	"errors"
	"fmt"

	"exemplos/livro2/capitulo-5/topico-6/projeto-1/banco"
	"exemplos/livro2/capitulo-5/topico-6/projeto-1/sdk"
)

func main() {
	var b banco.Banco = banco.NovoAdaptadorSDK(sdk.NovoCliente(map[string]float64{"ana": 100.10, "bia": 0}))
	ctx := context.Background()

	s, _ := b.Saldo(ctx, "ana")
	fmt.Println("saldo de ana (centavos):", s)

	op, err := b.Transferir(ctx, "ana", "bia", 2599)
	fmt.Println("transferencia:", op, err)

	_, err = b.Transferir(ctx, "ana", "bia", 50000)
	fmt.Println("saldo insuficiente?", errors.Is(err, banco.ErrSaldoInsuficiente), "|", err)

	_, err = b.Saldo(ctx, "zeca")
	fmt.Println("conta inexistente?", errors.Is(err, banco.ErrContaInexistente), "|", err)
}
