// Comando migrar: cadastra em um servidor em execução os endereços de um arquivo.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/migracao"
)

func main() {
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()
	err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return // o pacote flag já imprimiu o uso; pedir ajuda não é erro
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "migrar:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("migrar", flag.ContinueOnError)
	fs.SetOutput(stderr)
	servidor := fs.String("servidor", "http://127.0.0.1:8080", "endereço do servidor de links")
	arquivo := fs.String("arquivo", "-", "arquivo com um endereço por linha (- lê a entrada padrão)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	entrada := stdin
	if *arquivo != "-" {
		f, err := os.Open(*arquivo)
		if err != nil {
			return err
		}
		defer f.Close()
		entrada = f
	}

	res, err := migracao.Importar(ctx, migracao.Cliente{Base: *servidor}, entrada)
	fmt.Fprintf(stdout, "lidas=%d criadas=%d falhas=%d\n", res.Lidas, res.Criadas, len(res.Falhas))
	for _, f := range res.Falhas {
		fmt.Fprintf(stdout, "linha %d (%s): %s\n", f.Linha, f.URL, f.Erro)
	}
	if err != nil {
		return err
	}
	if len(res.Falhas) > 0 {
		return errors.New("algumas linhas falharam")
	}
	return nil
}
