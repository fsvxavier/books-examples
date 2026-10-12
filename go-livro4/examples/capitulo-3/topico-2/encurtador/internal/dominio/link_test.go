package dominio_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/dominio"
)

var instante = time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC)

func TestNovoLink(t *testing.T) {
	l, err := dominio.NovoLink("abc234", "  https://go.dev/doc/ ", instante)
	if err != nil {
		t.Fatal(err)
	}
	if l.Destino != "https://go.dev/doc/" || l.Acessos != 0 || !l.Criado.Equal(instante) {
		t.Errorf("link = %+v", l)
	}
}

func TestDestinosInvalidos(t *testing.T) {
	casos := map[string]string{
		"vazio":          "",
		"sem esquema":    "go.dev",
		"ftp":            "ftp://x.org/a",
		"javascript":     "javascript:alert(1)",
		"sem host":       "http:///a",
		"longo demais":   "http://x.org/" + strings.Repeat("a", 3000),
		"url malformada": "http://[::1",
	}
	for nome, destino := range casos {
		t.Run(nome, func(t *testing.T) {
			_, err := dominio.NovoLink("abc234", destino, instante)
			if !errors.Is(err, dominio.ErrDestinoInvalido) {
				t.Fatalf("erro = %v, esperado ErrDestinoInvalido", err)
			}
		})
	}
}

func TestMensagemSemEsquema(t *testing.T) {
	_, err := dominio.NormalizarDestino("go.dev")
	if err == nil || !strings.Contains(err.Error(), "faltou o esquema") {
		t.Fatalf("erro = %v", err)
	}
}

func TestCodigosInvalidos(t *testing.T) {
	for _, c := range []string{"", "abc23", "abc2345", "ABC234", "abc230", "abcdel"} {
		if err := dominio.ValidarCodigo(c); !errors.Is(err, dominio.ErrCodigoInvalido) {
			t.Errorf("ValidarCodigo(%q) = %v, esperado ErrCodigoInvalido", c, err)
		}
	}
	if err := dominio.ValidarCodigo("abc234"); err != nil {
		t.Errorf("abc234 deveria ser válido: %v", err)
	}
}
