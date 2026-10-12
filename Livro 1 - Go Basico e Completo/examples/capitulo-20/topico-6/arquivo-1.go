package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func rodar(maxOpen, maxIdle int, clientes int) {
	db, err := sql.Open("pgx", "postgres://livro:livro@127.0.0.1:55432/livro?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx := context.Background()
	inicio := time.Now()
	var wg sync.WaitGroup
	for range clientes {
		wg.Go(func() {
			var x int
			// cada consulta segura a conexao por 20ms (pg_sleep)
			if err := db.QueryRowContext(ctx, `SELECT 1 FROM pg_sleep(0.02)`).Scan(&x); err != nil {
				log.Println(err)
			}
		})
	}
	wg.Wait()
	s := db.Stats()
	fmt.Printf("max=%-3d idle=%-3d clientes=%d tempo=%-6s abertas=%d esperas=%d espera_total=%s fechadas_idle=%d\n",
		maxOpen, maxIdle, clientes, time.Since(inicio).Round(time.Millisecond),
		s.OpenConnections, s.WaitCount, s.WaitDuration.Round(time.Millisecond), s.MaxIdleClosed)
}

func main() {
	rodar(1, 1, 50)
	rodar(5, 5, 50)
	rodar(25, 25, 50)
	rodar(25, 2, 50)

	// timeout esperando conexao
	db, _ := sql.Open("pgx", "postgres://livro:livro@127.0.0.1:55432/livro?sslmode=disable")
	defer db.Close()
	db.SetMaxOpenConns(1)
	tx, _ := db.Begin() // segura a unica conexao
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err := db.QueryRowContext(ctx, `SELECT 1`).Scan(new(int))
	fmt.Println("sem conexao livre:", err)
	tx.Rollback()
}
