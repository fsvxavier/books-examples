// Package contagem é a fatia "contar acessos". Não tem rota: é uma reação ao
// evento LinkAcessado, inscrita no barramento pela composição.
package contagem

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/dominio"
)

// Manipulador devolve a reação que soma um ao total do código, em uma única
// instrução que o banco executa de forma atômica: o primeiro acesso cria a
// linha, os seguintes somam. Eventos de outros tipos são ignorados.
func Manipulador(pool *pgxpool.Pool) func(context.Context, dominio.Evento) error {
	return func(ctx context.Context, e dominio.Evento) error {
		acesso, ok := e.(dominio.LinkAcessado)
		if !ok {
			return nil
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO acessos (codigo, total) VALUES ($1, 1)
			 ON CONFLICT (codigo) DO UPDATE SET total = acessos.total + 1`, acesso.Codigo.String())
		return err
	}
}
