package barramento_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-6/topico-4/encurtador/internal/dominio"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-6/topico-4/encurtador/internal/infra/barramento"
)

func evento(t *testing.T) dominio.Evento {
	t.Helper()
	c, err := dominio.NovoCodigo("abc234")
	if err != nil {
		t.Fatal(err)
	}
	return dominio.LinkAcessado{Codigo: c, Em: time.Unix(0, 0)}
}

func TestEntregaEmOrdemATodos(t *testing.T) {
	b := barramento.Novo(slog.New(slog.DiscardHandler))
	var visto []string
	b.Inscrever(func(_ context.Context, e dominio.Evento) error {
		visto = append(visto, "primeiro:"+e.Nome())
		return nil
	})
	b.Inscrever(func(_ context.Context, e dominio.Evento) error {
		visto = append(visto, "segundo:"+e.Nome())
		return nil
	})

	b.Publicar(t.Context(), evento(t), evento(t))

	esperado := []string{
		"primeiro:link_acessado", "segundo:link_acessado",
		"primeiro:link_acessado", "segundo:link_acessado",
	}
	if !slices.Equal(visto, esperado) {
		t.Errorf("visto = %v", visto)
	}
}

func TestManipuladorQueFalhaNaoImpedeOsOutros(t *testing.T) {
	var buf bytes.Buffer
	b := barramento.Novo(slog.New(slog.NewTextHandler(&buf, nil)))
	chamado := false
	b.Inscrever(func(context.Context, dominio.Evento) error { return errors.New("a fila está fora do ar") })
	b.Inscrever(func(context.Context, dominio.Evento) error { chamado = true; return nil })

	b.Publicar(t.Context(), evento(t))

	if !chamado {
		t.Error("o segundo manipulador deveria ter rodado")
	}
	if !strings.Contains(buf.String(), "a fila está fora do ar") || !strings.Contains(buf.String(), "link_acessado") {
		t.Errorf("a falha deveria ter sido registrada, e nomear o evento: %q", buf.String())
	}
}

func TestSemInscritosNaoFazNada(t *testing.T) {
	barramento.Novo(slog.New(slog.DiscardHandler)).Publicar(t.Context(), evento(t))
}
