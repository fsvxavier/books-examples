// Package postgrestest dá a cada teste um esquema (schema) próprio no banco.
// Os pacotes de um módulo testam em paralelo, em processos diferentes; se todos
// usassem a mesma tabela, um TRUNCATE de um apagaria os dados do outro.
package postgrestest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// DSNIsolado cria um esquema novo e devolve o endereço do banco já apontando
// para ele (search_path). Ao fim do teste o esquema é apagado com tudo dentro.
// Se dsn for vazio, pula o teste.
func DSNIsolado(t testing.TB, dsn string) string {
	t.Helper()
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
