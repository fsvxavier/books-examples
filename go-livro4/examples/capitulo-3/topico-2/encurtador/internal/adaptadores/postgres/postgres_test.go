package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/adaptadores/postgres"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/aplicacao"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/aplicacao/repositoriotest"
)

// esquemaProprio cria um esquema (schema) novo no banco de teste e devolve o
// endereço já apontando para ele. Os pacotes de um módulo testam em paralelo,
// em processos diferentes; se todos usassem a tabela links do esquema padrão,
// o TRUNCATE de um apagaria os dados do outro.
func esquemaProprio(t *testing.T, dsn string) string {
	t.Helper()
	var sufixo [6]byte
	_, _ = rand.Read(sufixo[:])
	nome := "teste_" + hex.EncodeToString(sufixo[:])

	conn, err := pgx.Connect(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(t.Context())
	if _, err := conn.Exec(t.Context(), "CREATE SCHEMA "+nome); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// O t.Context() já foi cancelado quando a limpeza roda: contexto novo.
		ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		if c, err := pgx.Connect(ctx, dsn); err == nil {
			defer c.Close(ctx)
			_, _ = c.Exec(ctx, "DROP SCHEMA "+nome+" CASCADE")
		}
	})
	return dsn + "?search_path=" + nome
}

// O teste usa um PostgreSQL de verdade, indicado por ENCURTADOR_TESTE_DSN.
// Sem a variável, é pulado: a suíte continua verde em máquinas sem Docker.
func TestContrato(t *testing.T) {
	dsn := os.Getenv("ENCURTADOR_TESTE_DSN")
	if dsn == "" {
		t.Skip("defina ENCURTADOR_TESTE_DSN para testar contra um PostgreSQL")
	}
	pool, err := pgxpool.New(t.Context(), esquemaProprio(t, dsn))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	repositoriotest.Executar(t, func(t *testing.T) aplicacao.Repositorio {
		repo := postgres.Novo(pool)
		if err := repo.Migrar(t.Context()); err != nil {
			t.Fatal(err)
		}
		// Cada subteste começa com a tabela vazia (a do esquema deste teste).
		if _, err := pool.Exec(t.Context(), `TRUNCATE links`); err != nil {
			t.Fatal(err)
		}
		return repo
	})
}
