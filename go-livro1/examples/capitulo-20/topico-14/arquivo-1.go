package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Cliente struct {
	ID       int64
	Nome     string
	Email    string
	Telefone sql.Null[string]
	CriadoEm time.Time
}

func main() {
	ctx := context.Background()

	cfg := mysql.NewConfig()
	cfg.User = "livro"
	cfg.Passwd = "livro"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:53306"
	cfg.DBName = "livro"
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	cfg.Params = map[string]string{"charset": "utf8mb4"}
	cfg.MultiStatements = true
	fmt.Println("DSN:", cfg.FormatDSN())

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(3 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS clientes;
		CREATE TABLE clientes (
			id        BIGINT AUTO_INCREMENT PRIMARY KEY,
			nome      VARCHAR(100) NOT NULL,
			email     VARCHAR(150) NOT NULL UNIQUE,
			telefone  VARCHAR(20) NULL,
			visitas   INT NOT NULL DEFAULT 0,
			criado_em DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Fatal(err)
	}

	// Create: MySQL nao tem RETURNING, o id vem de LastInsertId
	res, err := db.ExecContext(ctx,
		`INSERT INTO clientes (nome, email) VALUES (?, ?)`, "Ana Souza", "ana@exemplo.com")
	if err != nil {
		log.Fatal(err)
	}
	id, _ := res.LastInsertId()
	fmt.Println("id:", id)

	db.ExecContext(ctx, `INSERT INTO clientes (nome, email, telefone) VALUES (?,?,?)`, "Bruno Lima", "bruno@exemplo.com", "11 99999-0000")

	// Upsert
	for range 3 {
		res, err = db.ExecContext(ctx,
			`INSERT INTO clientes (nome, email) VALUES (?, ?)
			 ON DUPLICATE KEY UPDATE visitas = visitas + 1`, "Ana Souza", "ana@exemplo.com")
		if err != nil {
			log.Fatal(err)
		}
		n, _ := res.RowsAffected()
		fmt.Println("upsert rowsAffected:", n)
	}

	// Read
	rows, err := db.QueryContext(ctx, `SELECT id, nome, email, telefone, criado_em FROM clientes ORDER BY id`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var c Cliente
		if err := rows.Scan(&c.ID, &c.Nome, &c.Email, &c.Telefone, &c.CriadoEm); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d %-10s %-18s %v %s\n", c.ID, c.Nome, c.Email, c.Telefone, c.CriadoEm.Format("15:04:05"))
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Update com valor igual: RowsAffected zero
	res, _ = db.ExecContext(ctx, `UPDATE clientes SET nome = ? WHERE id = ?`, "Ana Souza", id)
	n, _ := res.RowsAffected()
	fmt.Println("update sem mudanca:", n)

	// erro tipado
	_, err = db.ExecContext(ctx, `INSERT INTO clientes (nome, email) VALUES ('X', 'bruno@exemplo.com')`)
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		fmt.Printf("erro mysql numero=%d msg=%q\n", me.Number, me.Message)
	}

	// Delete
	res, _ = db.ExecContext(ctx, `DELETE FROM clientes WHERE id = ?`, id)
	n, _ = res.RowsAffected()
	fmt.Println("removidas:", n)
}
