package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Produto struct {
	ID       int64          `db:"id"`
	Nome     string         `db:"nome"`
	Preco    int64          `db:"preco"`
	Tags     []string       `db:"tags"`
	Extra    map[string]any `db:"extra"`
	Removido *time.Time     `db:"removido_em"`
}

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://livro:livro@127.0.0.1:55432/livro")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `
		DROP TABLE IF EXISTS itens;
		CREATE TABLE itens (
			id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			nome        TEXT NOT NULL UNIQUE,
			preco       BIGINT NOT NULL,
			tags        TEXT[] NOT NULL DEFAULT '{}',
			extra       JSONB NOT NULL DEFAULT '{}',
			removido_em TIMESTAMPTZ
		)`)
	if err != nil {
		log.Fatal(err)
	}

	// argumentos nomeados, array e jsonb sem conversao manual
	var id int64
	err = conn.QueryRow(ctx,
		`INSERT INTO itens (nome, preco, tags, extra)
		 VALUES (@nome, @preco, @tags, @extra) RETURNING id`,
		pgx.NamedArgs{
			"nome":  "Cadeira",
			"preco": 89900,
			"tags":  []string{"escritorio", "ergonomica"},
			"extra": map[string]any{"cor": "preta", "peso_kg": 11.5},
		}).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("id:", id)

	// Batch: varias instrucoes em uma ida e volta
	b := &pgx.Batch{}
	for i, n := range []string{"Mesa", "Luminaria", "Suporte"} {
		b.Queue(`INSERT INTO itens (nome, preco) VALUES ($1, $2)`, n, (i+1)*10000)
	}
	b.Queue(`SELECT count(*) FROM itens`).QueryRow(func(r pgx.Row) error {
		var n int
		if err := r.Scan(&n); err != nil {
			return err
		}
		fmt.Println("total apos batch:", n)
		return nil
	})
	if err := conn.SendBatch(ctx, b).Close(); err != nil {
		log.Fatal(err)
	}

	// CopyFrom: carga em massa pelo protocolo COPY
	linhas := make([][]any, 0, 1000)
	for i := range 1000 {
		linhas = append(linhas, []any{fmt.Sprintf("lote-%04d", i), int64(i)})
	}
	n, err := conn.CopyFrom(ctx, pgx.Identifier{"itens"}, []string{"nome", "preco"}, pgx.CopyFromRows(linhas))
	fmt.Println("copiadas:", n, err)

	// CollectRows + RowToStructByName
	rows, _ := conn.Query(ctx, `SELECT id, nome, preco, tags, extra, removido_em FROM itens WHERE id <= $1 ORDER BY id`, id+1)
	itens, err := pgx.CollectRows(rows, pgx.RowToStructByName[Produto])
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range itens {
		fmt.Printf("%+v\n", p)
	}

	// CollectOneRow e ErrNoRows
	rows, _ = conn.Query(ctx, `SELECT id, nome, preco, tags, extra, removido_em FROM itens WHERE id = $1`, -1)
	_, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[Produto])
	fmt.Println("sem linhas:", errors.Is(err, pgx.ErrNoRows), err)

	// erro tipado do PostgreSQL
	_, err = conn.Exec(ctx, `INSERT INTO itens (nome, preco) VALUES ('Mesa', 1)`)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		fmt.Printf("codigo=%s constraint=%s tabela=%s msg=%q\n", pgErr.Code, pgErr.ConstraintName, pgErr.TableName, pgErr.Message)
	}

	// o erro derruba a transacao implicita mas nao a conexao
	var total int
	conn.QueryRow(ctx, `SELECT count(*) FROM itens`).Scan(&total)
	fmt.Println("total final:", total)
}
