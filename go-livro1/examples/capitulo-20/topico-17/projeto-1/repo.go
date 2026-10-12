package e12

import (
	"context"
	"database/sql"
	"fmt"
)

type Repo struct{ db *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Nome(ctx context.Context, id int64) (string, error) {
	var nome string
	err := r.db.QueryRowContext(ctx, `SELECT nome FROM usuarios WHERE id = $1`, id).Scan(&nome)
	if err != nil {
		return "", fmt.Errorf("buscar usuario %d: %w", id, err)
	}
	return nome, nil
}
