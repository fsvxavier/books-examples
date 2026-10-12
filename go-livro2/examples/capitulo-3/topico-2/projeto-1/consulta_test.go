package consulta

import (
	"reflect"
	"strings"
	"testing"
)

func TestConstruir(t *testing.T) {
	q, err := De("produtos", "nome", "preco").
		Colunas("id", "nome").
		Onde("preco >= ?", 1000).
		OndeSe(true, "nome ILIKE ?", "%teclado%").
		OndeSe(false, "estoque > ?", 0). // filtro opcional ausente
		OrdenarPor("preco", true).
		Limite(20).Deslocar(40).
		Construir()
	if err != nil {
		t.Fatal(err)
	}
	wantSQL := "SELECT id, nome FROM produtos WHERE preco >= $1 AND nome ILIKE $2 ORDER BY preco DESC LIMIT $3 OFFSET $4"
	if q.SQL != wantSQL {
		t.Errorf("SQL:\n got %s\nwant %s", q.SQL, wantSQL)
	}
	if !reflect.DeepEqual(q.Args, []any{1000, "%teclado%", 20, 40}) {
		t.Errorf("Args = %v", q.Args)
	}
}

func TestEntradaMaliciosaNaoViraTextoSQL(t *testing.T) {
	q, err := De("usuarios").Onde("email = ?", "x' OR '1'='1").Construir()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(q.SQL, "OR") || q.Args[0] != "x' OR '1'='1" {
		t.Fatalf("o valor vazou para o SQL: %s %v", q.SQL, q.Args)
	}
}

func TestOrdenacaoForaDaListaERecusada(t *testing.T) {
	_, err := De("produtos", "nome").OrdenarPor("preco; DROP TABLE produtos", false).Construir()
	if err == nil || !strings.Contains(err.Error(), "nao permitido") {
		t.Fatalf("esperava recusa, veio %v", err)
	}
}

func TestErrosAcumulados(t *testing.T) {
	_, err := De("").Limite(0).Deslocar(-1).Onde("sem placeholder", 1).Construir()
	if err == nil {
		t.Fatal("esperava erros")
	}
	for _, trecho := range []string{"tabela obrigatoria", "limite 0", "deslocamento -1", "exatamente um"} {
		if !strings.Contains(err.Error(), trecho) {
			t.Errorf("o erro nao menciona %q:\n%v", trecho, err)
		}
	}
}

func TestDeslocamentoExigeLimite(t *testing.T) {
	if _, err := De("t").Deslocar(10).Construir(); err == nil {
		t.Fatal("esperava erro")
	}
}

func TestConsultaSimples(t *testing.T) {
	q, _ := De("t").Construir()
	if q.SQL != "SELECT * FROM t" || len(q.Args) != 0 {
		t.Fatalf("%+v", q)
	}
}
