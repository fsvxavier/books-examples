package test

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const modulo = "exemplos/livro4/capitulo-3/topico-2/encurtador/"

// dependencias devolve TODAS as dependências, diretas e indiretas, de um pacote
// do módulo. Olhar só os imports diretos deixaria passar um atalho como
// dominio -> util -> net/http.
func dependencias(t *testing.T, pacote string) []string {
	t.Helper()
	saida, err := exec.CommandContext(t.Context(), "go", "list", "-deps", "-f", "{{.ImportPath}}", modulo+pacote).CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps %s: %v\n%s", pacote, err, saida)
	}
	return strings.Fields(string(saida))
}

// alcanca diz se dep é o pacote proibido ou está dentro dele.
func alcanca(dep, proibido string) bool {
	return dep == proibido || strings.HasPrefix(dep, strings.TrimSuffix(proibido, "/")+"/")
}

// TestRegraDaDependencia confere que as dependências só apontam para dentro:
// cada pacote lista o que NÃO pode alcançar, nem por caminhos indiretos.
func TestRegraDaDependencia(t *testing.T) {
	infra := []string{"net/http", "database/sql", "github.com/", "encoding/json", "log/slog"}
	regras := []struct {
		pacote     string
		proibidos  []string
		observacao string
	}{
		{"internal/dominio", append([]string{
			modulo + "internal/aplicacao", modulo + "internal/adaptadores", modulo + "internal/contrato",
			modulo + "internal/app", modulo + "internal/config",
		}, infra...), "o domínio não conhece ninguém"},
		{"internal/aplicacao", append([]string{
			modulo + "internal/adaptadores", modulo + "internal/contrato", modulo + "internal/app",
		}, infra...), "a aplicação só conhece o domínio"},
		{"internal/adaptadores/web", []string{
			modulo + "internal/adaptadores/memoria", modulo + "internal/adaptadores/postgres", "github.com/jackc",
		}, "o adaptador de entrada não conhece os de saída"},
		{"internal/adaptadores/memoria", []string{
			modulo + "internal/adaptadores/web", modulo + "internal/adaptadores/postgres", "net/http",
		}, "adaptadores não se conhecem"},
		{"internal/adaptadores/postgres", []string{
			modulo + "internal/adaptadores/web", modulo + "internal/adaptadores/memoria", "net/http",
		}, "adaptadores não se conhecem"},
	}
	for _, r := range regras {
		t.Run(r.pacote, func(t *testing.T) {
			deps := dependencias(t, r.pacote)
			for _, p := range r.proibidos {
				if slices.ContainsFunc(deps, func(dep string) bool { return alcanca(dep, p) }) {
					t.Errorf("%s alcança %s, direta ou indiretamente (%s)", r.pacote, p, r.observacao)
				}
			}
		})
	}
}
