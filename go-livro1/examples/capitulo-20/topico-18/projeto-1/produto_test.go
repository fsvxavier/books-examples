package catalogo

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// contrato roda os MESMOS testes contra qualquer implementacao de Repositorio.
func contrato(t *testing.T, novo func(t *testing.T) Repositorio) {
	t.Run("criar e buscar", func(t *testing.T) {
		r, ctx := novo(t), t.Context()
		p := Produto{SKU: "A1", Nome: "Teclado", Preco: 25990, Tags: []string{"periferico", "usb"}}
		if err := r.Criar(ctx, p); err != nil {
			t.Fatal(err)
		}
		got, err := r.Buscar(ctx, "A1")
		if err != nil {
			t.Fatal(err)
		}
		if got.Nome != p.Nome || got.Preco != p.Preco || !slices.Equal(got.Tags, p.Tags) {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("sku duplicado", func(t *testing.T) {
		r, ctx := novo(t), t.Context()
		r.Criar(ctx, Produto{SKU: "A1", Nome: "x", Preco: 1})
		if err := r.Criar(ctx, Produto{SKU: "A1", Nome: "y", Preco: 2}); !errors.Is(err, ErrDuplicado) {
			t.Fatalf("esperava ErrDuplicado, veio %v", err)
		}
	})
	t.Run("nao encontrado", func(t *testing.T) {
		r, ctx := novo(t), t.Context()
		if _, err := r.Buscar(ctx, "?"); !errors.Is(err, ErrNaoEncontrado) {
			t.Fatalf("buscar: %v", err)
		}
		if err := r.AtualizarPreco(ctx, "?", 1); !errors.Is(err, ErrNaoEncontrado) {
			t.Fatalf("atualizar: %v", err)
		}
		if err := r.Remover(ctx, "?"); !errors.Is(err, ErrNaoEncontrado) {
			t.Fatalf("remover: %v", err)
		}
	})
	t.Run("atualizar com o mesmo preco nao e erro", func(t *testing.T) {
		r, ctx := novo(t), t.Context()
		r.Criar(ctx, Produto{SKU: "A1", Nome: "x", Preco: 100})
		if err := r.AtualizarPreco(ctx, "A1", 100); err != nil {
			t.Fatalf("mesmo preco: %v", err)
		}
		if err := r.AtualizarPreco(ctx, "A1", 150); err != nil {
			t.Fatal(err)
		}
		if got, _ := r.Buscar(ctx, "A1"); got.Preco != 150 {
			t.Fatalf("preco %d", got.Preco)
		}
	})
	t.Run("listar por tag e remover", func(t *testing.T) {
		r, ctx := novo(t), t.Context()
		r.Criar(ctx, Produto{SKU: "B", Nome: "b", Preco: 1, Tags: []string{"x", "y"}})
		r.Criar(ctx, Produto{SKU: "A", Nome: "a", Preco: 1, Tags: []string{"x"}})
		r.Criar(ctx, Produto{SKU: "C", Nome: "c", Preco: 1, Tags: []string{"y"}})
		lista, err := r.ListarPorTag(ctx, "x")
		if err != nil || len(lista) != 2 || lista[0].SKU != "A" || lista[1].SKU != "B" {
			t.Fatalf("lista %+v, %v", lista, err)
		}
		if err := r.Remover(ctx, "A"); err != nil {
			t.Fatal(err)
		}
		if lista, _ = r.ListarPorTag(ctx, "x"); len(lista) != 1 {
			t.Fatalf("apos remover: %+v", lista)
		}
	})
}

func TestPostgres(t *testing.T) {
	contrato(t, func(t *testing.T) Repositorio {
		pool, err := pgxpool.New(t.Context(), "postgres://livro:livro@127.0.0.1:55432/livro")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pool.Close)
		r, err := NovoPostgres(t.Context(), pool)
		if err != nil {
			t.Fatal(err)
		}
		return r
	})
}

func TestMongo(t *testing.T) {
	contrato(t, func(t *testing.T) Repositorio {
		c, err := mongo.Connect(options.Client().ApplyURI("mongodb://127.0.0.1:57017"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Disconnect(context.Background()) })
		r, err := NovoMongo(t.Context(), c.Database("livro").Collection("produtos_mongo"))
		if err != nil {
			t.Fatal(err)
		}
		return r
	})
}
