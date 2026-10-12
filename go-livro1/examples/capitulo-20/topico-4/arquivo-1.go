package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Produto struct {
	ID        int64
	Nome      string
	Preco     int64
	Descricao sql.Null[string]
	CriadoEm  time.Time
}

var ErrNaoEncontrado = errors.New("produto nao encontrado")

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", "postgres://livro:livro@127.0.0.1:55432/livro?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS produtos;
		CREATE TABLE produtos (
			id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			nome      TEXT NOT NULL UNIQUE,
			preco     BIGINT NOT NULL CHECK (preco >= 0),
			descricao TEXT,
			criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		log.Fatal(err)
	}

	// Create
	var id int64
	err = db.QueryRowContext(ctx,
		`INSERT INTO produtos (nome, preco) VALUES ($1, $2) RETURNING id`,
		"Teclado", 25990).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("criado id:", id)

	res, err := db.ExecContext(ctx,
		`INSERT INTO produtos (nome, preco, descricao) VALUES ($1,$2,$3), ($4,$5,$6)`,
		"Mouse", 8990, "Sem fio", "Monitor", 129900, nil)
	if err != nil {
		log.Fatal(err)
	}
	n, _ := res.RowsAffected()
	fmt.Println("linhas inseridas:", n)
	if _, err := res.LastInsertId(); err != nil {
		fmt.Println("LastInsertId:", err)
	}

	// Read
	rows, err := db.QueryContext(ctx,
		`SELECT id, nome, preco, descricao, criado_em FROM produtos ORDER BY id`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p Produto
		if err := rows.Scan(&p.ID, &p.Nome, &p.Preco, &p.Descricao, &p.CriadoEm); err != nil {
			log.Fatal(err)
		}
		desc := "(sem descricao)"
		if p.Descricao.Valid {
			desc = p.Descricao.V
		}
		fmt.Printf("%d %-8s %7d %s %s\n", p.ID, p.Nome, p.Preco, desc, p.CriadoEm.UTC().Format("15:04"))
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// ErrNoRows
	var p Produto
	err = db.QueryRowContext(ctx, `SELECT id, nome FROM produtos WHERE id = $1`, 999).Scan(&p.ID, &p.Nome)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNaoEncontrado
	}
	fmt.Println("busca 999:", err)

	// Update
	res, err = db.ExecContext(ctx, `UPDATE produtos SET preco = preco * 90 / 100 WHERE nome = $1`, "Monitor")
	if err != nil {
		log.Fatal(err)
	}
	n, _ = res.RowsAffected()
	fmt.Println("atualizadas:", n)

	// Delete
	res, err = db.ExecContext(ctx, `DELETE FROM produtos WHERE id = $1`, id)
	if err != nil {
		log.Fatal(err)
	}
	n, _ = res.RowsAffected()
	fmt.Println("removidas:", n)

	// violacao de constraint
	_, err = db.ExecContext(ctx, `INSERT INTO produtos (nome, preco) VALUES ('Mouse', 1)`)
	fmt.Println("duplicado:", err)
}
