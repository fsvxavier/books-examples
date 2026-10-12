package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrSaldo = errors.New("saldo insuficiente")

// ComTx executa fn dentro de uma transacao. Commit se fn retorna nil,
// Rollback em qualquer outro caso (inclusive panic).
func ComTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, fn func(*sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = errors.Join(err, rbErr)
			}
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func serializacao(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "40001"
}

// ComRetry repete a transacao quando o banco aborta por conflito de serializacao.
func ComRetry(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	opts := &sql.TxOptions{Isolation: sql.LevelSerializable}
	for tentativa := 1; ; tentativa++ {
		err := ComTx(ctx, db, opts, fn)
		if err == nil || !serializacao(err) || tentativa == 5 {
			return err
		}
		select {
		case <-time.After(time.Duration(tentativa) * 10 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func transferir(tx *sql.Tx, de, para int, valor int64) error {
	var saldo int64
	if err := tx.QueryRow(`SELECT saldo FROM contas WHERE id = $1`, de).Scan(&saldo); err != nil {
		return err
	}
	if saldo < valor {
		return ErrSaldo
	}
	if _, err := tx.Exec(`UPDATE contas SET saldo = saldo - $1 WHERE id = $2`, valor, de); err != nil {
		return err
	}
	_, err := tx.Exec(`UPDATE contas SET saldo = saldo + $1 WHERE id = $2`, valor, para)
	return err
}

// transferirTravando bloqueia as duas linhas (sempre em ordem de id, para
// evitar deadlock) antes de ler. Quem chega depois espera, em vez de falhar.
func transferirTravando(tx *sql.Tx, de, para int, valor int64) error {
	if _, err := tx.Exec(`SELECT 1 FROM contas WHERE id IN ($1, $2) ORDER BY id FOR UPDATE`, de, para); err != nil {
		return err
	}
	return transferir(tx, de, para, valor)
}

func saldos(ctx context.Context, db *sql.DB) string {
	var a, b, soma int64
	_ = db.QueryRowContext(ctx, `SELECT
		(SELECT saldo FROM contas WHERE id=1),
		(SELECT saldo FROM contas WHERE id=2),
		(SELECT sum(saldo) FROM contas)`).Scan(&a, &b, &soma)
	return fmt.Sprintf("conta1=%d conta2=%d soma=%d", a, b, soma)
}

func main() {
	ctx := context.Background()
	db, err := sql.Open("pgx", "postgres://livro:livro@127.0.0.1:55432/livro?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.ExecContext(ctx, `DROP TABLE IF EXISTS contas`)
	db.ExecContext(ctx, `CREATE TABLE contas (id INT PRIMARY KEY, saldo BIGINT NOT NULL CHECK (saldo >= 0))`)
	db.ExecContext(ctx, `INSERT INTO contas VALUES (1, 1000), (2, 500)`)
	fmt.Println("inicio:  ", saldos(ctx, db))

	err = ComTx(ctx, db, nil, func(tx *sql.Tx) error { return transferir(tx, 1, 2, 300) })
	fmt.Println("300 ok:  ", err, saldos(ctx, db))

	err = ComTx(ctx, db, nil, func(tx *sql.Tx) error { return transferir(tx, 1, 2, 5000) })
	fmt.Println("5000:    ", err, "| ErrSaldo:", errors.Is(err, ErrSaldo), saldos(ctx, db))

	// falha no meio: debita e depois quebra. O rollback desfaz o debito.
	err = ComTx(ctx, db, nil, func(tx *sql.Tx) error {
		if _, err := tx.Exec(`UPDATE contas SET saldo = saldo - 100 WHERE id = 1`); err != nil {
			return err
		}
		return errors.New("falha simulada")
	})
	fmt.Println("meio:    ", err, saldos(ctx, db))

	// concorrencia: 20 transferencias de 10 com SERIALIZABLE
	var wg sync.WaitGroup
	var mu sync.Mutex
	falhas := 0
	for range 20 {
		wg.Go(func() {
			err := ComRetry(ctx, db, func(tx *sql.Tx) error { return transferir(tx, 1, 2, 10) })
			if err != nil {
				mu.Lock()
				falhas++
				mu.Unlock()
				fmt.Println("erro apos 5 tentativas:", err)
			}
		})
	}
	wg.Wait()
	fmt.Println("serializable:", falhas, "falhas;", saldos(ctx, db))

	// mesma carga, agora com FOR UPDATE no nivel padrao (READ COMMITTED)
	falhas = 0
	for range 20 {
		wg.Go(func() {
			err := ComTx(ctx, db, nil, func(tx *sql.Tx) error { return transferirTravando(tx, 1, 2, 10) })
			if err != nil {
				mu.Lock()
				falhas++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	fmt.Println("for update:  ", falhas, "falhas;", saldos(ctx, db))
}
