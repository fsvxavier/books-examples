// Package postgres implementa nucleo.Repositorio e nucleo.Contador sobre o
// PostgreSQL, com pgx v5 e um pool de conexões criado por quem monta o programa.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-5/topico-3/encurtador/internal/nucleo"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-5/topico-3/encurtador/internal/nucleo/dominio"
)

// codigoUnicidade é o SQLSTATE de violação de chave única.
const codigoUnicidade = "23505"

// Duas tabelas, porque são dois conceitos: o agregado Link (que quase não
// muda) e o contador de acessos (que muda a cada visita).
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

// Armazenamento guarda links e contadores no PostgreSQL.
type Armazenamento struct {
	pool *pgxpool.Pool
}

var _ nucleo.Armazenamento = (*Armazenamento)(nil)

// Novo recebe o pool pronto: quem o abre também decide quando fechá-lo.
func Novo(pool *pgxpool.Pool) *Armazenamento {
	return &Armazenamento{pool: pool}
}

// Migrar cria as tabelas, se ainda não existirem.
func (a *Armazenamento) Migrar(ctx context.Context) error {
	if _, err := a.pool.Exec(ctx, esquema); err != nil {
		return fmt.Errorf("criando as tabelas: %w", err)
	}
	return nil
}

// Salvar insere o link. Uma violação da chave única vira dominio.ErrCodigoEmUso:
// a decisão é do banco, então não existe corrida entre verificar e gravar.
// O agregado é lido pelos métodos públicos e gravado em colunas simples.
func (a *Armazenamento) Salvar(ctx context.Context, l dominio.Link) error {
	var expira *time.Time // NULL quando o link não tem prazo
	if fim, tem := l.Expira(); tem {
		expira = &fim
	}
	_, err := a.pool.Exec(ctx,
		`INSERT INTO links (codigo, destino, criado, expira) VALUES ($1, $2, $3, $4)`,
		l.Codigo().String(), l.Destino().String(), l.Criado(), expira)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codigoUnicidade {
		return dominio.ErrCodigoEmUso
	}
	return err
}

// Buscar lê as colunas e reconstitui o agregado, reaplicando as regras dos
// objetos de valor: dados corrompidos no banco não viram um Link inválido.
func (a *Armazenamento) Buscar(ctx context.Context, codigo dominio.Codigo) (dominio.Link, error) {
	var cod, destino string
	var criado time.Time
	var expira *time.Time
	err := a.pool.QueryRow(ctx,
		`SELECT codigo, destino, criado, expira FROM links WHERE codigo = $1`, codigo.String(),
	).Scan(&cod, &destino, &criado, &expira)
	if errors.Is(err, pgx.ErrNoRows) {
		return dominio.Link{}, dominio.ErrNaoEncontrado
	}
	if err != nil {
		return dominio.Link{}, err
	}
	var fim time.Time
	if expira != nil {
		fim = *expira
	}
	return dominio.Reconstituir(cod, destino, criado, fim)
}

// Incrementar soma um acesso em uma única instrução, que o banco executa de
// forma atômica: o primeiro acesso cria a linha, os seguintes somam.
func (a *Armazenamento) Incrementar(ctx context.Context, codigo dominio.Codigo) error {
	_, err := a.pool.Exec(ctx,
		`INSERT INTO acessos (codigo, total) VALUES ($1, 1)
		 ON CONFLICT (codigo) DO UPDATE SET total = acessos.total + 1`, codigo.String())
	return err
}

// Total devolve os acessos do código; sem linha, zero.
func (a *Armazenamento) Total(ctx context.Context, codigo dominio.Codigo) (int, error) {
	var total int
	err := a.pool.QueryRow(ctx, `SELECT total FROM acessos WHERE codigo = $1`, codigo.String()).Scan(&total)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return total, err
}
