// Comando encurtador: o servidor de links curtos.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"exemplos/livro4/capitulo-2/topico-5/encurtador/internal/app"
)

func main() {
	err := app.Executar(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return // o pacote flag já imprimiu o uso; pedir ajuda não é erro
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "encurtador:", err)
		os.Exit(1)
	}
}
