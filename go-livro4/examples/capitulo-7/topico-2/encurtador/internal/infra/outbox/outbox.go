// Package outbox implementa o padrão da caixa de saída transacional: a
// mensagem é gravada NA MESMA TRANSAÇÃO que a mudança de estado, e um relay a
// publica depois. Não existe janela em que o estado mudou e a mensagem se perdeu.
package outbox

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-7/topico-2/encurtador/internal/infra/mensageria"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-7/topico-2/encurtador/internal/mensagens"
)

// Executor é o que Gravar precisa: executar um comando. Tanto uma transação
// (pgx.Tx) quanto o pool o satisfazem; para o padrão valer, passe a transação.
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Gravar insere a mensagem na caixa de saída.
func Gravar(ctx context.Context, ex Executor, topico string, env mensagens.Envelope) error {
	payload, err := env.Bytes()
	if err != nil {
		return err
	}
	_, err = ex.Exec(ctx,
		`INSERT INTO outbox (id, topico, payload) VALUES ($1, $2, $3)`, env.ID, topico, payload)
	return err
}

// Relay lê a caixa de saída e publica no Pub/Sub.
type Relay struct {
	Pool      *pgxpool.Pool
	Publisher message.Publisher
	Log       *slog.Logger
	Lote      int           // mensagens por rodada
	Intervalo time.Duration // espera quando não há nada a publicar
}

// Uma publica, no máximo, um lote de mensagens pendentes e devolve quantas
// foram publicadas. Cada mensagem é marcada como publicada logo depois de o
// Publisher aceitá-la; se uma publicação falhar, as anteriores ficam marcadas, e
// a rodada devolve o erro. Várias instâncias do relay podem rodar ao mesmo
// tempo: o FOR UPDATE SKIP LOCKED faz cada uma pegar mensagens diferentes.
func (r *Relay) Uma(ctx context.Context) (int, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) // desfaz, se não houve Commit

	linhas, err := tx.Query(ctx,
		`SELECT id, topico, payload FROM outbox
		  WHERE publicado IS NULL ORDER BY criado
		  LIMIT $1 FOR UPDATE SKIP LOCKED`, r.Lote)
	if err != nil {
		return 0, err
	}
	type pendente struct {
		id, topico string
		payload    []byte
	}
	pendentes, err := pgx.CollectRows(linhas, func(l pgx.CollectableRow) (p pendente, err error) {
		err = l.Scan(&p.id, &p.topico, &p.payload)
		return p, err
	})
	if err != nil {
		return 0, err
	}

	publicadas := 0
	var falha error
	for _, p := range pendentes {
		if err := r.Publisher.Publish(p.topico, message.NewMessage(p.id, p.payload)); err != nil {
			falha = fmt.Errorf("publicando %s: %w", p.id, err)
			break
		}
		if _, err := tx.Exec(ctx, `UPDATE outbox SET publicado = now() WHERE id = $1`, p.id); err != nil {
			return publicadas, err
		}
		publicadas++
	}
	if err := tx.Commit(ctx); err != nil {
		// A transação não valeu: as mensagens já publicadas voltam a ficar
		// pendentes e serão publicadas de novo. É a entrega "pelo menos uma vez".
		return 0, fmt.Errorf("confirmando a rodada: %w", err)
	}
	return publicadas, falha
}

// Executar roda o relay até o contexto ser cancelado.
func (r *Relay) Executar(ctx context.Context) error {
	for {
		n, err := r.Uma(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			r.Log.Error("relay da caixa de saída", "erro", mensageria.Resumo(err))
		}
		if n > 0 && err == nil {
			continue // havia trabalho: tenta de novo na hora
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.Intervalo):
		}
	}
}
