// Package postgres implementa nucleo.Repositorio sobre o PostgreSQL, com
// pgx v5 e um pool de conexões criado por quem monta o programa.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

// codigoUnicidade é o SQLSTATE de violação de chave única.
const codigoUnicidade = "23505"

const esquema = `
CREATE TABLE IF NOT EXISTS links (
	codigo  text        PRIMARY KEY,
	destino text        NOT NULL,
	criado  timestamptz NOT NULL,
	acessos bigint      NOT NULL DEFAULT 0
)`

// Repositorio guarda os links em uma tabela do PostgreSQL.
type Repositorio struct {
	pool *pgxpool.Pool
}

// Novo recebe o pool pronto: quem o abre também decide quando fechá-lo.
func Novo(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{pool: pool}
}

// Migrar cria a tabela, se ainda não existir.
func (r *Repositorio) Migrar(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, esquema); err != nil {
		return fmt.Errorf("criando a tabela links: %w", err)
	}
	return nil
}

// Salvar insere o link. Uma violação da chave única vira dominio.ErrCodigoEmUso,
// a decisão é do banco, então não existe corrida entre verificar e gravar.
func (r *Repositorio) Salvar(ctx context.Context, l dominio.Link) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO links (codigo, destino, criado, acessos) VALUES ($1, $2, $3, $4)`,
		l.Codigo, l.Destino, l.Criado, l.Acessos)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codigoUnicidade {
		return dominio.ErrCodigoEmUso
	}
	return err
}

// Buscar devolve o link, ou dominio.ErrNaoEncontrado.
func (r *Repositorio) Buscar(ctx context.Context, codigo string) (dominio.Link, error) {
	return r.uma(ctx,
		`SELECT codigo, destino, criado, acessos FROM links WHERE codigo = $1`, codigo)
}

// RegistrarAcesso soma um acesso em uma única instrução UPDATE ... RETURNING,
// que o banco executa de forma atômica.
func (r *Repositorio) RegistrarAcesso(ctx context.Context, codigo string) (dominio.Link, error) {
	return r.uma(ctx,
		`UPDATE links SET acessos = acessos + 1 WHERE codigo = $1
		 RETURNING codigo, destino, criado, acessos`, codigo)
}

func (r *Repositorio) uma(ctx context.Context, sql string, args ...any) (dominio.Link, error) {
	var l dominio.Link
	err := r.pool.QueryRow(ctx, sql, args...).Scan(&l.Codigo, &l.Destino, &l.Criado, &l.Acessos)
	if errors.Is(err, pgx.ErrNoRows) {
		return dominio.Link{}, dominio.ErrNaoEncontrado
	}
	return l, err
}
