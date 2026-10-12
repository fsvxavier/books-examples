// Package app monta e executa o servidor: é o que cmd/encurtador chama.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/armazem"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/config"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/servidor"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-5/encurtador/internal/versao"
)

// Executar é o programa de verdade. Tudo que vem de fora (contexto,
// argumentos, ambiente, saídas) chega por parâmetro, então um teste consegue
// executá-lo por inteiro sem tocar no processo.
func Executar(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) error {
	ctx, parar := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer parar()

	cfg, err := config.Ler(args, getenv, stderr)
	if err != nil {
		return err
	}
	if cfg.MostrarVersao {
		fmt.Fprintln(stdout, versao.Texto())
		return nil
	}
	log := slog.New(slog.NewTextHandler(stderr, nil))

	ln, err := net.Listen("tcp", cfg.Endereco)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           servidor.Novo(cfg.BaseURL, armazem.Novo(), log),
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Fprintf(stdout, "ouvindo em %s\n", ln.Addr())

	falha := make(chan error, 1)
	go func() { falha <- srv.Serve(ln) }()

	select {
	case err := <-falha:
		return err
	case <-ctx.Done():
	}

	// O prazo do desligamento não pode herdar o cancelamento que nos trouxe aqui.
	prazo, cancelar := context.WithTimeout(context.WithoutCancel(ctx), cfg.Desligamento)
	defer cancelar()
	if err := srv.Shutdown(prazo); err != nil {
		return fmt.Errorf("desligamento: %w", err)
	}
	if err := <-falha; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	fmt.Fprintln(stdout, "desligado")
	return nil
}
