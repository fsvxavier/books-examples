package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func leitura(ctx context.Context, iso pgx.TxIsoLevel) {
	a, _ := pgx.Connect(ctx, "postgres://livro:livro@127.0.0.1:55432/livro")
	b, _ := pgx.Connect(ctx, "postgres://livro:livro@127.0.0.1:55432/livro")
	defer a.Close(ctx)
	defer b.Close(ctx)

	a.Exec(ctx, `DROP TABLE IF EXISTS preco; CREATE TABLE preco (v INT); INSERT INTO preco VALUES (100)`)

	tx, err := a.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso})
	if err != nil {
		log.Fatal(err)
	}
	var v1, v2 int
	tx.QueryRow(ctx, `SELECT v FROM preco`).Scan(&v1)

	b.Exec(ctx, `UPDATE preco SET v = 200`) // outra sessao confirma no meio

	tx.QueryRow(ctx, `SELECT v FROM preco`).Scan(&v2)
	tx.Commit(ctx)
	fmt.Printf("%-16s primeira leitura=%d segunda leitura=%d\n", iso, v1, v2)
}

func main() {
	ctx := context.Background()
	leitura(ctx, pgx.ReadCommitted)
	leitura(ctx, pgx.RepeatableRead)
}
