// Package catalogo define o contrato de um repositorio de produtos e duas
// implementacoes dele: PostgreSQL (pgx) e MongoDB.
package catalogo

import (
	"context"
	"errors"
)

var (
	ErrNaoEncontrado = errors.New("produto nao encontrado")
	ErrDuplicado     = errors.New("sku ja existe")
)

type Produto struct {
	SKU   string
	Nome  string
	Preco int64
	Tags  []string
}

type Repositorio interface {
	Criar(ctx context.Context, p Produto) error
	Buscar(ctx context.Context, sku string) (Produto, error)
	AtualizarPreco(ctx context.Context, sku string, preco int64) error
	Remover(ctx context.Context, sku string) error
	ListarPorTag(ctx context.Context, tag string) ([]Produto, error)
}
