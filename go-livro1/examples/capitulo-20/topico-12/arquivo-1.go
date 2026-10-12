package loja

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migracoes/*.sql
var migracoes embed.FS

// Migrar aplica, em ordem, cada arquivo ainda nao registrado em schema_migrations.
// O lock consultivo impede que duas instancias migrem ao mesmo tempo.
func Migrar(ctx context.Context, pool *pgxpool.Pool) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(727274)`); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
			versao TEXT PRIMARY KEY, aplicada_em TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
			return err
		}
		arquivos, err := fs.Glob(migracoes, "migracoes/*.sql")
		if err != nil {
			return err
		}
		sort.Strings(arquivos)
		for _, a := range arquivos {
			var existe bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE versao = $1)`, a).Scan(&existe); err != nil {
				return err
			}
			if existe {
				continue
			}
			sqlTexto, err := migracoes.ReadFile(a)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, string(sqlTexto)); err != nil {
				return fmt.Errorf("migracao %s: %w", a, err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (versao) VALUES ($1)`, a); err != nil {
				return err
			}
		}
		return nil
	})
}
