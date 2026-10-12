package auditoria_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/adaptadores/saida/auditoria"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo"
	"exemplos/livro4/capitulo-4/topico-2/encurtador/internal/nucleo/dominio"
)

// Garantias de compilação: os dois tipos satisfazem a porta.
var (
	_ nucleo.Auditoria = (*auditoria.Slog)(nil)
	_ nucleo.Auditoria = auditoria.Descartar{}
)

func TestSlogEscreveUmRegistroPorEvento(t *testing.T) {
	var buf bytes.Buffer
	a := auditoria.NovoSlog(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: func(grupos []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.Attr{} // sem hora, para a saída ser comparável
			}
			return attr
		},
	})))
	l := dominio.Link{Codigo: "abc234", Destino: "https://go.dev", Criado: time.Unix(0, 0), Acessos: 3}

	a.LinkCriado(t.Context(), l)
	a.LinkAcessado(t.Context(), l)

	linhas := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(linhas) != 2 {
		t.Fatalf("%d registros, esperados 2:\n%s", len(linhas), buf.String())
	}
	var r map[string]any
	if err := json.Unmarshal([]byte(linhas[1]), &r); err != nil {
		t.Fatal(err)
	}
	if r["evento"] != "link_acessado" || r["codigo"] != "abc234" || r["acessos"] != float64(3) {
		t.Errorf("registro = %v", r)
	}
}
