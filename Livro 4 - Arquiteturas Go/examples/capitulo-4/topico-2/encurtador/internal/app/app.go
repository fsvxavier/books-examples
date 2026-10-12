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

	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/adaptadores/entrada/web"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/adaptadores/saida/auditoria"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/adaptadores/saida/memoria"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/adaptadores/saida/postgres"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/config"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/versao"
)

// NovoNucleo liga o hexágono às suas portas de saída: o repositório recebido,
// o gerador e o relógio de produção e a auditoria em log. O resultado é a
// PORTA DE ENTRADA, e é ela que qualquer adaptador primário recebe.
func NovoNucleo(repo nucleo.Repositorio, log *slog.Logger) nucleo.Encurtador {
	return nucleo.NovoServico(repo, nucleo.CodigoAleatorio, time.Now, auditoria.NovoSlog(log))
}

// NovoHandler liga o adaptador HTTP à porta de entrada.
func NovoHandler(baseURL string, enc nucleo.Encurtador, log *slog.Logger) http.Handler {
	return web.Novo(baseURL, enc, log)
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

	repo, fechar, err := AbrirRepositorio(ctx, cfg.DSN, log)
	if err != nil {
		return err
	}
	defer fechar()

	ln, err := net.Listen("tcp", cfg.Endereco)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           NovoHandler(cfg.BaseURL, NovoNucleo(repo, log), log),
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

// AbrirRepositorio escolhe a implementação da porta com base na configuração.
// É a única linha de decisão: o resto do programa não sabe qual foi escolhida.
func AbrirRepositorio(ctx context.Context, dsn string, log *slog.Logger) (nucleo.Repositorio, func(), error) {
	if dsn == "" {
		log.Info("guardando os links em memória")
		return memoria.Novo(), func() {}, nil
	}
	pool, err := pgxpool.New(ctx, dsn)
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
