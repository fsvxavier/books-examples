// Package app é a raiz de composição: o único lugar que conhece todas as
// camadas e decide quem implementa cada porta.
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

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/adaptadores/memoria"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/adaptadores/postgres"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/adaptadores/web"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/aplicacao"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/config"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/versao"
)

// NovoHandler liga as camadas: o repositório recebido entra no serviço, e o
// serviço entra no adaptador HTTP. Testes usam esta função para montar o
// programa com o repositório que quiserem.
func NovoHandler(baseURL string, repo aplicacao.Repositorio, log *slog.Logger) http.Handler {
	casos := aplicacao.NovoServico(repo, aplicacao.CodigoAleatorio, time.Now)
	return web.Novo(baseURL, casos, log)
}

// Executar é o programa de verdade. Tudo que vem de fora (contexto,
// argumentos, ambiente, saídas) chega por parâmetro.
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

	repo, fechar, err := abrirRepositorio(ctx, cfg, log)
	if err != nil {
		return err
	}
	defer fechar()

	ln, err := net.Listen("tcp", cfg.Endereco)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           NovoHandler(cfg.BaseURL, repo, log),
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

// abrirRepositorio escolhe a implementação da porta com base na configuração.
// É a única linha de decisão: o resto do programa não sabe qual foi escolhida.
func abrirRepositorio(ctx context.Context, cfg config.Config, log *slog.Logger) (aplicacao.Repositorio, func(), error) {
	if cfg.DSN == "" {
		log.Info("guardando os links em memória")
		return memoria.Novo(), func() {}, nil
	}
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("abrindo o pool do PostgreSQL: %w", err)
	}
	repo := postgres.Novo(pool)
	if err := repo.Migrar(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}
	log.Info("guardando os links no PostgreSQL")
	return repo, pool.Close, nil
}
