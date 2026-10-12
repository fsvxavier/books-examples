package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func novoPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig("postgres://livro:livro@127.0.0.1:55432/livro")
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 8
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnLifetimeJitter = 2 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = time.Minute
	cfg.ConnConfig.RuntimeParams["application_name"] = "livro-cap20"

	// roda uma vez para cada conexao nova
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, `SET statement_timeout = '5s'`)
		return err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

func main() {
	ctx := context.Background()
	pool, err := novoPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	st := pool.Stat()
	fmt.Printf("inicio: total=%d idle=%d max=%d\n", st.TotalConns(), st.IdleConns(), st.MaxConns())

	var app, timeout string
	pool.QueryRow(ctx, `SELECT current_setting('application_name'), current_setting('statement_timeout')`).Scan(&app, &timeout)
	fmt.Println("sessao:", app, timeout)

	// carga concorrente
	var wg sync.WaitGroup
	inicio := time.Now()
	for range 100 {
		wg.Go(func() {
			var x int
			pool.QueryRow(ctx, `SELECT 1 FROM pg_sleep(0.02)`).Scan(&x)
		})
	}
	wg.Wait()
	st = pool.Stat()
	fmt.Printf("100 consultas: %s total=%d adquiridas=%d espera_vazia=%d espera=%s\n",
		time.Since(inicio).Round(time.Millisecond), st.TotalConns(), st.AcquireCount(),
		st.EmptyAcquireCount(), st.AcquireDuration().Round(time.Millisecond))

	// Acquire manual: varias instrucoes na MESMA conexao
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}
	conn.Exec(ctx, `SET search_path TO public`)
	var pid int
	conn.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid)
	var pid2 int
	conn.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid2)
	fmt.Println("mesma conexao:", pid == pid2)
	conn.Release() // devolve ao pool, nao fecha

	// transacao com helper: commit se retorna nil, rollback caso contrario
	pool.Exec(ctx, `DROP TABLE IF EXISTS saldos; CREATE TABLE saldos (id INT PRIMARY KEY, v BIGINT NOT NULL)`)
	pool.Exec(ctx, `INSERT INTO saldos VALUES (1, 100)`)
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE saldos SET v = v + 50 WHERE id = 1`); err != nil {
			return err
		}
		return errors.New("desfaz")
	})
	var v int
	pool.QueryRow(ctx, `SELECT v FROM saldos WHERE id = 1`).Scan(&v)
	fmt.Println("rollback:", err, "v =", v)

	err = pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE saldos SET v = v + 50 WHERE id = 1`)
		return err
	})
	pool.QueryRow(ctx, `SELECT v FROM saldos WHERE id = 1`).Scan(&v)
	fmt.Println("commit:", err, "v =", v)

	// statement_timeout da sessao aborta consulta lenta
	err = pool.QueryRow(ctx, `SELECT pg_sleep(6)`).Scan(new(any))
	fmt.Println("lenta:", err)

	// timeout do contexto cancela no servidor
	cctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	err = pool.QueryRow(cctx, `SELECT pg_sleep(3)`).Scan(new(any))
	fmt.Println("contexto:", err)

	// o pool continua saudavel
	err = pool.Ping(ctx)
	fmt.Println("ping apos timeout:", err)
}
