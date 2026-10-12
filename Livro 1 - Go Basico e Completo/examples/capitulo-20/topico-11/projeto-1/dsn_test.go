package e06

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const dsn = "postgres://livro:livro@127.0.0.1:55432/livro?sslmode=disable"

type linha struct {
	ID    int64
	Nome  string
	Preco int64
}

func BenchmarkPgxPool(b *testing.B) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn+"&pool_max_conns=16")
	if err != nil {
		b.Fatal(err)
	}
	defer pool.Close()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var l linha
		for pb.Next() {
			err := pool.QueryRow(ctx, `SELECT id, nome, preco FROM itens WHERE id = $1`, 1).Scan(&l.ID, &l.Nome, &l.Preco)
			if err != nil {
				b.Error(err)
				return
			}
		}
	})
}

func BenchmarkDatabaseSQL(b *testing.B) {
	ctx := context.Background()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(16)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var l linha
		for pb.Next() {
			err := db.QueryRowContext(ctx, `SELECT id, nome, preco FROM itens WHERE id = $1`, 1).Scan(&l.ID, &l.Nome, &l.Preco)
			if err != nil {
				b.Error(err)
				return
			}
		}
	})
}
