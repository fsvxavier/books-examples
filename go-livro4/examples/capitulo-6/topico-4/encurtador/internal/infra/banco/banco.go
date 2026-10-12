// Package banco abre o pool do PostgreSQL e cria as tabelas. É a infraestrutura
// que as fatias compartilham: cada fatia escreve o SQL que lhe interessa, mas o
// esquema (quem existe, com que colunas) tem um dono só, aqui.
package banco

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// esquema cria as duas tabelas do sistema. links é lida pelas fatias criar,
// seguir e consultar; acessos é escrita por contagem e lida por consultar.
const esquema = `
CREATE TABLE IF NOT EXISTS links (
	codigo  text        PRIMARY KEY,
	destino text        NOT NULL,
	criado  timestamptz NOT NULL,
	expira  timestamptz
);
CREATE TABLE IF NOT EXISTS acessos (
	codigo text   PRIMARY KEY,
	total  bigint NOT NULL
)`

// Abrir conecta ao PostgreSQL e confere a conexão.
func Abrir(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("abrindo o pool do PostgreSQL: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectando ao PostgreSQL: %w", err)
	}
	return pool, nil
}

// Migrar cria as tabelas, se ainda não existirem.
func Migrar(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, esquema); err != nil {
		return fmt.Errorf("criando as tabelas: %w", err)
	}
	return nil
}
