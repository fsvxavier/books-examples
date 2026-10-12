package longo

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// EsquemaEntidades cria o grafo de entidades e relações descobertas na conversa.
const EsquemaEntidades = `
CREATE TABLE IF NOT EXISTS entidades (
	id      bigserial PRIMARY KEY,
	usuario text   NOT NULL,
	nome    text   NOT NULL,
	tipo    text   NOT NULL,
	UNIQUE (usuario, nome)
);
CREATE TABLE IF NOT EXISTS relacoes (
	usuario text   NOT NULL,
	de      bigint NOT NULL REFERENCES entidades(id) ON DELETE CASCADE,
	para    bigint NOT NULL REFERENCES entidades(id) ON DELETE CASCADE,
	tipo    text   NOT NULL,
	PRIMARY KEY (usuario, de, para, tipo)
);
`

// Aresta é uma relação encontrada na vizinhança de uma entidade.
type Aresta struct {
	De, Tipo, Para string
	Salto          int
}

// Entidade cria (ou encontra) uma entidade e devolve o seu id.
func (l *Loja) Entidade(ctx context.Context, usuario, nome, tipo string) (id int64, err error) {
	err = l.comUsuario(ctx, usuario, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			INSERT INTO entidades (usuario, nome, tipo) VALUES ($1, $2, $3)
			ON CONFLICT (usuario, nome) DO UPDATE SET tipo = EXCLUDED.tipo
			RETURNING id`, usuario, nome, tipo).Scan(&id)
	})
	return id, err
}

// Relacionar liga duas entidades; repetir a mesma relação não faz nada.
func (l *Loja) Relacionar(ctx context.Context, usuario string, de int64, tipo string, para int64) error {
	return l.comUsuario(ctx, usuario, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO relacoes (usuario, de, para, tipo) VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING`, usuario, de, para, tipo)
		return err
	})
}

// Vizinhanca devolve as relações a até saltos passos da entidade, nos dois
// sentidos, sem entrar em ciclos (a cláusula CYCLE do PostgreSQL cuida disso).
func (l *Loja) Vizinhanca(ctx context.Context, usuario, nome string, saltos int) (arestas []Aresta, err error) {
	err = l.comUsuario(ctx, usuario, func(tx pgx.Tx) error {
		linhas, err := tx.Query(ctx, `
			WITH RECURSIVE andar(entidade, salto) AS (
				SELECT id, 0 FROM entidades WHERE usuario = $1 AND nome = $2
				UNION ALL
				SELECT CASE WHEN r.de = a.entidade THEN r.para ELSE r.de END, a.salto + 1
				FROM andar a
				JOIN relacoes r ON r.usuario = $1 AND (r.de = a.entidade OR r.para = a.entidade)
				WHERE a.salto < $3
			) CYCLE entidade SET ciclo USING caminho
			SELECT d.nome, r.tipo, p.nome, MIN(a.salto) + 1 AS salto
			FROM andar a
			JOIN relacoes r ON r.usuario = $1 AND (r.de = a.entidade OR r.para = a.entidade)
			JOIN entidades d ON d.id = r.de
			JOIN entidades p ON p.id = r.para
			WHERE NOT a.ciclo AND a.salto < $3
			GROUP BY d.nome, r.tipo, p.nome
			ORDER BY salto, d.nome, r.tipo, p.nome`, usuario, nome, saltos)
		if err != nil {
			return err
		}
		arestas, err = pgx.CollectRows(linhas, func(r pgx.CollectableRow) (Aresta, error) {
			var a Aresta
			err := r.Scan(&a.De, &a.Tipo, &a.Para, &a.Salto)
			return a, err
		})
		return err
	})
	return arestas, err
}
