// Package app é a raiz de composição: monta as fatias sobre a infraestrutura
// compartilhada e executa o servidor.
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

	"github.com/jackc/pgx/v5/pgxpool"

	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/auditoria"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/config"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/consultar"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/contagem"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/criar"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/infra/banco"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/infra/barramento"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/seguir"
	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/versao"
)

// NovoHandler liga as três fatias, o barramento e a contagem de acessos sobre
// o pool recebido. Os testes de ponta a ponta usam esta função diretamente.
func NovoHandler(pool *pgxpool.Pool, baseURL string, validade time.Duration, log *slog.Logger) http.Handler {
	bus := barramento.Novo(log)
	bus.Inscrever(auditoria.Manipulador(log))
	bus.Inscrever(contagem.Manipulador(pool))

	mux := http.NewServeMux()
	criar.Novo(pool, bus, criar.CodigoAleatorio, time.Now, validade).Registrar(mux, baseURL, log)
	seguir.Novo(pool, bus, time.Now).Registrar(mux, log)
	consultar.Novo(pool).Registrar(mux, log)
	mux.HandleFunc("GET /saude", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintln(w, `{"estado":"ok"}`)
	})
	return mux
}

// Executar é o programa de verdade. Tudo que vem de fora (contexto,
// argumentos, ambiente, saídas) chega por parâmetro. Diferente das versões em
// memória, esta exige um PostgreSQL (ENCURTADOR_DSN ou -dsn).
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
	if cfg.DSN == "" {
		return errors.New("informe o PostgreSQL em ENCURTADOR_DSN ou -dsn")
	}
	log := slog.New(slog.NewTextHandler(stderr, nil))

	pool, err := banco.Abrir(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := banco.Migrar(ctx, pool); err != nil {
		return err
	}

	ln, err := net.Listen("tcp", cfg.Endereco)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           NovoHandler(pool, cfg.BaseURL, 0, log),
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
