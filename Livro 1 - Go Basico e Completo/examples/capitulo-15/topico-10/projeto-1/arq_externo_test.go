package arq_test // pacote EXTERNO: só enxerga a API pública, como um usuário de verdade

import (
	"os"
	"slices"
	"strings"
	"testing"

	"exemplo.com/arq"
)

func TestEmail(t *testing.T) {
	got, err := arq.Email("Ana", 42)
	if err != nil || !strings.Contains(got, "Ana") || !strings.Contains(got, "42") {
		t.Errorf("Email = %q, %v", got, err)
	}
}

func TestMigracoes(t *testing.T) {
	got, err := arq.Migracoes()
	quer := []string{"assets/sql/001_pedidos.sql", "assets/sql/002_valor.sql"}
	if err != nil || !slices.Equal(got, quer) {
		t.Errorf("Migracoes = %v, %v", got, err)
	}
}

func TestDadosDeTeste(t *testing.T) {
	b, err := os.ReadFile("testdata/caso1.json")
	if err != nil || len(b) == 0 {
		t.Fatal(err)
	}
}
