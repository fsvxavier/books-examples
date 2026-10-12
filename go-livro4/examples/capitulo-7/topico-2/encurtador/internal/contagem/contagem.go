// Package contagem é a fatia "contar acessos": um CONSUMIDOR da mensagem
// link_acessado. Não tem rota; quem a aciona é o broker.
package contagem

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-7/topico-2/encurtador/internal/mensagens"
)

// Consumidor identifica esta fatia na tabela processadas: cada consumidor tem
// a sua própria memória das mensagens que já tratou.
const Consumidor = "contagem"

// Tratar processa UMA mensagem e é idempotente: tratar a mesma mensagem duas
// vezes soma um acesso só. A entrega do broker é "pelo menos uma vez", então a
// mesma mensagem pode chegar repetida (a queda de um consumidor, o relay que
// publicou e não conseguiu confirmar). O identificador da mensagem vai para a
// tabela processadas na MESMA transação que soma o acesso: ou os dois valem, ou
// nenhum.
func Tratar(ctx context.Context, pool *pgxpool.Pool, msg *message.Message) error {
	env, err := mensagens.DoBytes(msg.Payload)
	if err != nil {
		return err
	}
	acesso, err := mensagens.Decodificar[mensagens.LinkAcessadoV1](env)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	novo, err := tx.Exec(ctx,
		`INSERT INTO processadas (consumidor, id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		Consumidor, env.ID)
	if err != nil {
		return fmt.Errorf("registrando a mensagem %s: %w", env.ID, err)
	}
	if novo.RowsAffected() == 0 {
		return nil // já tratada: ignora, e o Ack acontece normalmente
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO acessos (codigo, total) VALUES ($1, 1)
		 ON CONFLICT (codigo) DO UPDATE SET total = acessos.total + 1`, acesso.Codigo); err != nil {
		return fmt.Errorf("somando o acesso: %w", err)
	}
	return tx.Commit(ctx)
}

// Registrar inscreve o consumidor no roteador, no tópico dos acessos. Se
// Tratar devolver erro, o Watermill faz Nack, e a mensagem é reentregue.
func Registrar(r *message.Router, sub message.Subscriber, pool *pgxpool.Pool) {
	r.AddConsumerHandler("contagem", mensagens.TopicoLinkAcessado, sub, func(msg *message.Message) error {
		return Tratar(msg.Context(), pool, msg)
	})
}
