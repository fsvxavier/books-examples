package logs

import (
	"reflect"
	"testing"
)

var versoes = []struct {
	nome string
	f    func([]string) Relatorio
}{
	{"v0", ResumirV0},
	{"v1", ResumirV1},
	{"v2", ResumirV2},
	{"v3", ResumirV3},
	{"v4", ResumirV4},
}

func TestAmostra(t *testing.T) {
	linhas := []string{
		"2026-10-05T12:00:01Z GET /api/produtos/123 200 12.5ms usuario=ana",
		"2026-10-05T12:00:02Z GET /api/produtos/9 500 80.0ms usuario=bruno",
		"2026-10-05T12:00:03Z POST /api/pedidos/5/itens/7 201 30.2ms usuario=ana",
		"lixo",
	}
	for _, v := range versoes {
		t.Run(v.nome, func(t *testing.T) {
			got := v.f(linhas)
			quer := Relatorio{
				Total: 3, Erros: 1, P95Ms: 80, MaisAtivo: "ana",
				PorRota: map[string]int{
					"GET /api/produtos/:id":           2,
					"POST /api/pedidos/:id/itens/:id": 1,
				},
			}
			if !reflect.DeepEqual(got, quer) {
				t.Errorf("got %+v\nquer %+v", got, quer)
			}
		})
	}
}

// Teste diferencial: todas as versoes tem de concordar com a V0 (a referencia).
func TestVersoesConcordam(t *testing.T) {
	linhas := Gerar(20000)
	ref := ResumirV0(linhas)
	for _, v := range versoes[1:] {
		if got := v.f(linhas); !reflect.DeepEqual(got, ref) {
			t.Errorf("%s difere da V0:\n got %+v\nquer %+v", v.nome, got.P95Ms, ref.P95Ms)
		}
	}
}

func BenchmarkResumir(b *testing.B) {
	linhas := Gerar(20000)
	for _, v := range versoes {
		b.Run("impl="+v.nome, func(b *testing.B) {
			for b.Loop() {
				v.f(linhas)
			}
		})
	}
}
