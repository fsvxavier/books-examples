package e14

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// naoCompila confirma que um pacote de exemplo NÃO compila, e que o erro diz o que se espera.
// É o teste de uma constraint: o que ela recusa faz parte do contrato, tanto quanto o que ela aceita.
func naoCompila(t *testing.T, pacote, trecho string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", os.DevNull, pacote)
	cmd.Dir = "../erros"
	saida, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("%s compilou, mas deveria ser recusado", pacote)
	}
	if !strings.Contains(string(saida), trecho) {
		t.Fatalf("%s: o erro não menciona %q:\n%s", pacote, trecho, saida)
	}
}

func TestAConstraintRecusaOQueDeveRecusar(t *testing.T) {
	t.Run("método ausente na interseção", func(t *testing.T) { naoCompila(t, "./t1", "missing method String") })
	t.Run("conjunto vazio", func(t *testing.T) { naoCompila(t, "./t2", "empty type set") })
	t.Run("ponteiro com método", func(t *testing.T) { naoCompila(t, "./p1", "has pointer receiver") })
	t.Run("constraint recursiva", func(t *testing.T) { naoCompila(t, "./r1", "wrong type for method Comparar") })
	t.Run("produto sem identificador", func(t *testing.T) { naoCompila(t, "./b1", "missing method ObterID") })
}
