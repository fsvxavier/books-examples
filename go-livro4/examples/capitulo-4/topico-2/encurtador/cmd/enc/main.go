// Comando enc: o encurtador de links na linha de comando, sem servidor.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/adaptadores/entrada/cli"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/app"
)

func main() {
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()
	err := run(ctx, os.Args[1:], os.Getenv, os.Stdin, os.Stdout, os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "enc:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("enc", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dsn := fs.String("dsn", getenv("ENCURTADOR_DSN"), "endereço do PostgreSQL (ou ENCURTADOR_DSN)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dsn == "" {
		// Um processo que termina em segundos não tem o que fazer com a memória.
		return errors.New("é preciso um PostgreSQL: defina ENCURTADOR_DSN ou use -dsn")
	}

	// Os registros de log vão para a saída de erro; a saída padrão fica só com o resultado.
	log := slog.New(slog.NewTextHandler(stderr, nil))
	repo, fechar, err := app.AbrirRepositorio(ctx, *dsn, log)
	if err != nil {
		return err
	}
	defer fechar()

	return cli.Executar(ctx, fs.Args(), app.NovoNucleo(repo, log), stdin, stdout, stderr)
}
