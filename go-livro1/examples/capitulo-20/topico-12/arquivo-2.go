package loja

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNaoEncontrado       = errors.New("nao encontrado")
	ErrEstoqueInsuficiente = errors.New("estoque insuficiente")
	ErrDuplicado           = errors.New("registro duplicado")
)

type Produto struct {
	ID      int64  `db:"id"`
	SKU     string `db:"sku"`
	Nome    string `db:"nome"`
	Preco   int64  `db:"preco"`
	Estoque int    `db:"estoque"`
}

type ItemPedido struct {
	ProdutoID int64 `db:"produto_id"`
	Qtd       int   `db:"qtd"`
	Preco     int64 `db:"preco"`
}

type Pedido struct {
	ID       int64        `db:"id"`
	Cliente  string       `db:"cliente"`
	Total    int64        `db:"total"`
	CriadoEm time.Time    `db:"criado_em"`
	Itens    []ItemPedido `db:"-"`
}

type Repo struct{ pool *pgxpool.Pool }

func NovoRepo(p *pgxpool.Pool) *Repo { return &Repo{pool: p} }

func traduz(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNaoEncontrado
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: %s", ErrDuplicado, pgErr.ConstraintName)
	}
	return err
}

func (r *Repo) CriarProduto(ctx context.Context, p Produto) (Produto, error) {
	rows, _ := r.pool.Query(ctx,
		`INSERT INTO produtos (sku, nome, preco, estoque)
		 VALUES (@sku, @nome, @preco, @estoque)
		 RETURNING id, sku, nome, preco, estoque`,
		pgx.NamedArgs{"sku": p.SKU, "nome": p.Nome, "preco": p.Preco, "estoque": p.Estoque})
	out, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Produto])
	return out, traduz(err)
}

func (r *Repo) BuscarProduto(ctx context.Context, id int64) (Produto, error) {
	rows, _ := r.pool.Query(ctx, `SELECT id, sku, nome, preco, estoque FROM produtos WHERE id = $1`, id)
	out, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Produto])
	return out, traduz(err)
}

// CriarPedido grava pedido e itens e baixa o estoque, tudo em uma transacao.
func (r *Repo) CriarPedido(ctx context.Context, cliente string, itens []ItemPedido) (Pedido, error) {
	var ped Pedido
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var total int64
		for i, it := range itens {
			// UPDATE condicional: baixa so se houver estoque, e devolve o preco atual
			var preco int64
			err := tx.QueryRow(ctx,
				`UPDATE produtos SET estoque = estoque - $2
				 WHERE id = $1 AND estoque >= $2 RETURNING preco`, it.ProdutoID, it.Qtd).Scan(&preco)
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("produto %d: %w", it.ProdutoID, ErrEstoqueInsuficiente)
			}
			if err != nil {
				return err
			}
			itens[i].Preco = preco
			total += preco * int64(it.Qtd)
		}
		err := tx.QueryRow(ctx,
			`INSERT INTO pedidos (cliente, total) VALUES ($1, $2) RETURNING id, criado_em`,
			cliente, total).Scan(&ped.ID, &ped.CriadoEm)
		if err != nil {
			return err
		}
		linhas := make([][]any, len(itens))
		for i, it := range itens {
			linhas[i] = []any{ped.ID, it.ProdutoID, it.Qtd, it.Preco}
		}
		_, err = tx.CopyFrom(ctx, pgx.Identifier{"itens_pedido"},
			[]string{"pedido_id", "produto_id", "qtd", "preco"}, pgx.CopyFromRows(linhas))
		ped.Cliente, ped.Total, ped.Itens = cliente, total, itens
		return err
	})
	return ped, err
}

// Pagina usa paginacao por keyset: nao depende de OFFSET, entao o custo
// nao cresce com a profundidade da pagina.
type Cursor struct {
	CriadoEm time.Time
	ID       int64
}

func (r *Repo) ListarPedidos(ctx context.Context, depois *Cursor, limite int) ([]Pedido, *Cursor, error) {
	var rows pgx.Rows
	if depois == nil {
		rows, _ = r.pool.Query(ctx,
			`SELECT id, cliente, total, criado_em FROM pedidos
			 ORDER BY criado_em DESC, id DESC LIMIT $1`, limite)
	} else {
		rows, _ = r.pool.Query(ctx,
			`SELECT id, cliente, total, criado_em FROM pedidos
			 WHERE (criado_em, id) < ($1, $2)
			 ORDER BY criado_em DESC, id DESC LIMIT $3`, depois.CriadoEm, depois.ID, limite)
	}
	peds, err := pgx.CollectRows(rows, pgx.RowToStructByName[Pedido])
	if err != nil || len(peds) < limite {
		return peds, nil, err
	}
	u := peds[len(peds)-1]
	return peds, &Cursor{CriadoEm: u.CriadoEm, ID: u.ID}, nil
}
