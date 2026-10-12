package catalogo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct{ pool *pgxpool.Pool }

func NovoPostgres(ctx context.Context, pool *pgxpool.Pool) (*Postgres, error) {
	_, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS produtos_pg;
		CREATE TABLE produtos_pg (
			sku   TEXT PRIMARY KEY,
			nome  TEXT NOT NULL,
			preco BIGINT NOT NULL,
			tags  TEXT[] NOT NULL DEFAULT '{}')`)
	if err != nil {
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (r *Postgres) Criar(ctx context.Context, p Produto) error {
	if p.Tags == nil {
		p.Tags = []string{} // slice nil vira NULL, e a coluna e NOT NULL
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO produtos_pg (sku, nome, preco, tags) VALUES ($1, $2, $3, $4)`,
		p.SKU, p.Nome, p.Preco, p.Tags)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicado
	}
	return err
}

func (r *Postgres) Buscar(ctx context.Context, sku string) (Produto, error) {
	var p Produto
	err := r.pool.QueryRow(ctx,
		`SELECT sku, nome, preco, tags FROM produtos_pg WHERE sku = $1`, sku).
		Scan(&p.SKU, &p.Nome, &p.Preco, &p.Tags)
	if errors.Is(err, pgx.ErrNoRows) {
		return Produto{}, ErrNaoEncontrado
	}
	return p, err
}

func (r *Postgres) AtualizarPreco(ctx context.Context, sku string, preco int64) error {
	tag, err := r.pool.Exec(ctx, `UPDATE produtos_pg SET preco = $2 WHERE sku = $1`, sku, preco)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Postgres) Remover(ctx context.Context, sku string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM produtos_pg WHERE sku = $1`, sku)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Postgres) ListarPorTag(ctx context.Context, tag string) ([]Produto, error) {
	rows, _ := r.pool.Query(ctx,
		`SELECT sku, nome, preco, tags FROM produtos_pg WHERE $1 = ANY(tags) ORDER BY sku`, tag)
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Produto, error) {
		var p Produto
		err := row.Scan(&p.SKU, &p.Nome, &p.Preco, &p.Tags)
		return p, err
	})
}
