// Package bancotest dá a cada teste um esquema (schema) próprio no PostgreSQL,
// já com as tabelas criadas. Os pacotes de um módulo testam em paralelo, em
// processos diferentes; se todos usassem as mesmas tabelas, o TRUNCATE de um
// apagaria os dados do outro.
package bancotest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-6/topico-4/encurtador/internal/infra/banco"
)

// DSNIsolado cria um esquema novo e devolve o endereço do banco já apontando
// para ele (search_path). Ao fim do teste o esquema é apagado com tudo dentro.
// Sem ENCURTADOR_TESTE_DSN, o teste é pulado.
func DSNIsolado(t testing.TB) string {
	t.Helper()
	dsn := os.Getenv("ENCURTADOR_TESTE_DSN")
	if dsn == "" {
		t.Skip("defina ENCURTADOR_TESTE_DSN para testar contra um PostgreSQL")
	}
	var sufixo [6]byte
	_, _ = rand.Read(sufixo[:])
	esquema := "teste_" + hex.EncodeToString(sufixo[:])

	conn, err := pgx.Connect(t.Context(), dsn)
	if err != nil {
		t.Fatalf("conectando ao PostgreSQL de teste: %v", err)
	}
	if _, err := conn.Exec(t.Context(), "CREATE SCHEMA "+esquema); err != nil {
		conn.Close(t.Context())
		t.Fatalf("criando o esquema %s: %v", esquema, err)
	}
	conn.Close(t.Context())

	t.Cleanup(func() {
		// O t.Context() é cancelado ANTES das funções de limpeza rodarem, então
		// a limpeza precisa de um contexto próprio.
		ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			t.Errorf("limpando o esquema %s: %v", esquema, err)
			return
		}
		defer conn.Close(ctx)
		if _, err := conn.Exec(ctx, "DROP SCHEMA "+esquema+" CASCADE"); err != nil {
			t.Errorf("apagando o esquema %s: %v", esquema, err)
		}
	})
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "search_path=" + esquema
}

// Pool devolve um pool sobre um esquema próprio, com as tabelas criadas.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	pool, err := banco.Abrir(t.Context(), DSNIsolado(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := banco.Migrar(t.Context(), pool); err != nil {
		t.Fatal(err)
	}
	return pool
}
