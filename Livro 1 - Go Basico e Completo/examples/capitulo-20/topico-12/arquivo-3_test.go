package loja

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// cada teste usa um schema proprio, entao os testes rodam em paralelo sem se enxergar
func novoPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := t.Context()
	base, err := pgxpool.New(ctx, "postgres://livro:livro@127.0.0.1:55432/livro")
	if err != nil {
		t.Fatal(err)
	}
	schema := "t_" + sanitiza(t.Name())
	if _, err := base.Exec(ctx, `DROP SCHEMA IF EXISTS `+schema+` CASCADE; CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	cfg := base.Config()
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		base.Exec(context.Background(), `DROP SCHEMA `+schema+` CASCADE`)
		base.Close()
	})
	if err := Migrar(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func sanitiza(s string) string {
	b := []byte(s)
	for i, c := range b {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			b[i] = '_'
		}
	}
	return string(b)
}

func TestCriarProdutoDuplicado(t *testing.T) {
	t.Parallel()
	r := NovoRepo(novoPool(t))
	if _, err := r.CriarProduto(t.Context(), Produto{SKU: "A", Nome: "A", Preco: 1, Estoque: 1}); err != nil {
		t.Fatal(err)
	}
	_, err := r.CriarProduto(t.Context(), Produto{SKU: "A", Nome: "B", Preco: 2, Estoque: 1})
	if !errors.Is(err, ErrDuplicado) {
		t.Fatalf("esperava ErrDuplicado, veio %v", err)
	}
}

func TestCriarPedidoRollbackSemEstoque(t *testing.T) {
	t.Parallel()
	r := NovoRepo(novoPool(t))
	ctx := t.Context()
	a, _ := r.CriarProduto(ctx, Produto{SKU: "A", Nome: "A", Preco: 100, Estoque: 5})
	b, _ := r.CriarProduto(ctx, Produto{SKU: "B", Nome: "B", Preco: 200, Estoque: 1})

	_, err := r.CriarPedido(ctx, "ana", []ItemPedido{{ProdutoID: a.ID, Qtd: 2}, {ProdutoID: b.ID, Qtd: 3}})
	if !errors.Is(err, ErrEstoqueInsuficiente) {
		t.Fatalf("esperava ErrEstoqueInsuficiente, veio %v", err)
	}
	// o estoque de A nao pode ter sido baixado
	got, _ := r.BuscarProduto(ctx, a.ID)
	if got.Estoque != 5 {
		t.Fatalf("rollback falhou: estoque de A = %d", got.Estoque)
	}
}

func TestCriarPedidoConcorrenteNaoVendeAMais(t *testing.T) {
	t.Parallel()
	r := NovoRepo(novoPool(t))
	ctx := t.Context()
	p, _ := r.CriarProduto(ctx, Produto{SKU: "RARO", Nome: "Raro", Preco: 100, Estoque: 10})

	var wg sync.WaitGroup
	var mu sync.Mutex
	vendidos, recusados := 0, 0
	for range 30 {
		wg.Go(func() {
			_, err := r.CriarPedido(ctx, "x", []ItemPedido{{ProdutoID: p.ID, Qtd: 1}})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				vendidos++
			case errors.Is(err, ErrEstoqueInsuficiente):
				recusados++
			default:
				t.Error(err)
			}
		})
	}
	wg.Wait()
	got, _ := r.BuscarProduto(ctx, p.ID)
	t.Logf("vendidos=%d recusados=%d estoque=%d", vendidos, recusados, got.Estoque)
	if vendidos != 10 || recusados != 20 || got.Estoque != 0 {
		t.Fatal("estoque inconsistente")
	}
}

func TestPaginacaoKeyset(t *testing.T) {
	t.Parallel()
	r := NovoRepo(novoPool(t))
	ctx := t.Context()
	p, _ := r.CriarProduto(ctx, Produto{SKU: "P", Nome: "P", Preco: 1, Estoque: 1000})
	for range 25 {
		if _, err := r.CriarPedido(ctx, "c", []ItemPedido{{ProdutoID: p.ID, Qtd: 1}}); err != nil {
			t.Fatal(err)
		}
	}
	var cur *Cursor
	vistos := map[int64]bool{}
	paginas := 0
	for {
		peds, prox, err := r.ListarPedidos(ctx, cur, 10)
		if err != nil {
			t.Fatal(err)
		}
		paginas++
		for _, p := range peds {
			if vistos[p.ID] {
				t.Fatalf("pedido %d repetido", p.ID)
			}
			vistos[p.ID] = true
		}
		if prox == nil {
			break
		}
		cur = prox
	}
	if len(vistos) != 25 || paginas != 3 {
		t.Fatalf("vistos=%d paginas=%d", len(vistos), paginas)
	}
}
