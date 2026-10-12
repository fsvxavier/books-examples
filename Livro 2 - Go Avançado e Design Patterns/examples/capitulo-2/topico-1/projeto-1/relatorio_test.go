package relatorio

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

var linhas = []Linha{{"teclado", 25990}, {"mouse", 8990}}

func TestFormatosReais(t *testing.T) {
	casos := []struct {
		nome string
		f    Fabrica
		want string
	}{
		{"csv", NovoCSV, "produto,total\nteclado,25990\nmouse,8990\n"},
		{"jsonl", NovoJSONL, "{\"Produto\":\"teclado\",\"Total\":25990}\n{\"Produto\":\"mouse\",\"Total\":8990}\n"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			var buf bytes.Buffer
			if err := (Servico{c.f}).Exportar(&buf, linhas); err != nil {
				t.Fatal(err)
			}
			if buf.String() != c.want {
				t.Fatalf("got %q\nwant %q", buf.String(), c.want)
			}
		})
	}
}

// Um exportador FALSO, entregue pela fabrica: testa o Servico sem formato nenhum.
type falso struct {
	escritas []string
	falhaEm  string
	fechou   bool
}

var errBoom = errors.New("boom")

func (f *falso) Escrever(l Linha) error {
	if l.Produto == f.falhaEm {
		return errBoom
	}
	f.escritas = append(f.escritas, l.Produto)
	return nil
}
func (f *falso) Fechar() error { f.fechou = true; return nil }

func TestServicoComFabricaFalsa(t *testing.T) {
	f := &falso{falhaEm: "mouse"}
	s := Servico{Nova: func(_ io.Writer) Exportador { return f }}

	err := s.Exportar(&bytes.Buffer{}, linhas)
	if !errors.Is(err, errBoom) || !strings.Contains(err.Error(), `"mouse"`) {
		t.Fatalf("esperava boom com o produto no erro, veio %v", err)
	}
	if !f.fechou {
		t.Fatal("o exportador devia ser fechado mesmo com erro")
	}
	if len(f.escritas) != 1 || f.escritas[0] != "teclado" {
		t.Fatalf("escritas: %v", f.escritas)
	}
}
