// Comando encurtador: um servidor de links curtos em um único pacote.
package main

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
)

func main() {
	err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "encurtador:", err)
		os.Exit(1)
	}
}

// run é o programa de verdade. Tudo que vem de fora (contexto, argumentos,
// ambiente, saídas) chega por parâmetro, então um teste consegue executá-lo
// por inteiro sem tocar no processo.
func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) error {
	ctx, parar := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer parar()

	cfg, err := lerConfig(args, getenv, stderr)
	if err != nil {
		return err
	}
	log := slog.New(slog.NewTextHandler(stderr, nil))

	ln, err := net.Listen("tcp", cfg.endereco)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           novoServidor(cfg, log),
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
	prazo, cancelar := context.WithTimeout(context.WithoutCancel(ctx), cfg.desligamento)
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
