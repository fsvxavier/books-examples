package web_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/adaptadores/web"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/aplicacao"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/contrato"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-3/topico-2/encurtador/internal/dominio"
)

// casosFalsos deixa cada teste decidir o que os casos de uso respondem.
// O adaptador HTTP é testado sem repositório, sem sorteio e sem relógio.
type casosFalsos struct {
	criar     func(ctx context.Context, destino string) (dominio.Link, error)
	resolver  func(ctx context.Context, codigo string) (dominio.Link, error)
	consultar func(ctx context.Context, codigo string) (dominio.Link, error)
}

func (c casosFalsos) Criar(ctx context.Context, d string) (dominio.Link, error) {
	return c.criar(ctx, d)
}

func (c casosFalsos) Resolver(ctx context.Context, k string) (dominio.Link, error) {
	return c.resolver(ctx, k)
}

func (c casosFalsos) Consultar(ctx context.Context, k string) (dominio.Link, error) {
	return c.consultar(ctx, k)
}

func chamar(casos web.Casos, metodo, alvo, corpo string) *httptest.ResponseRecorder {
	s := web.Novo("http://curto.test/", casos, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(metodo, alvo, strings.NewReader(corpo)))
	return rec
}

func TestTraducaoDeErrosParaStatus(t *testing.T) {
	desconhecido := errors.New("conexão com o banco caiu")
	casos := []struct {
		nome   string
		erro   error
		status int
	}{
		{"destino inválido", fmt.Errorf("%w: sem host", dominio.ErrDestinoInvalido), http.StatusUnprocessableEntity},
		{"não encontrado", dominio.ErrNaoEncontrado, http.StatusNotFound},
		{"código inválido", dominio.ErrCodigoInvalido, http.StatusNotFound},
		{"sem código livre", aplicacao.ErrSemCodigoLivre, http.StatusServiceUnavailable},
		{"erro desconhecido", desconhecido, http.StatusInternalServerError},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			falha := func(context.Context, string) (dominio.Link, error) { return dominio.Link{}, c.erro }
			rec := chamar(casosFalsos{criar: falha, resolver: falha, consultar: falha}, "POST", "/links", `{"url":"https://a.test"}`)
			if rec.Code != c.status {
				t.Errorf("status %d, esperado %d (%s)", rec.Code, c.status, rec.Body)
			}
			if strings.Contains(rec.Body.String(), desconhecido.Error()) {
				t.Errorf("o erro interno vazou para o cliente: %s", rec.Body)
			}
		})
	}
}

func TestCriarMontaOEnderecoCurto(t *testing.T) {
	casos := casosFalsos{criar: func(_ context.Context, d string) (dominio.Link, error) {
		if d != "https://go.dev/doc/" {
			t.Errorf("o caso de uso recebeu %q", d)
		}
		return dominio.Link{Codigo: "abc234", Destino: d}, nil
	}}
	rec := chamar(casos, "POST", "/links", `{"url":"https://go.dev/doc/"}`)
	var r contrato.RespostaCriar
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated || r.Curto != "http://curto.test/abc234" {
		t.Fatalf("status %d, resposta %+v", rec.Code, r)
	}
}

func TestRedirecionarEConsultar(t *testing.T) {
	link := dominio.Link{Codigo: "abc234", Destino: "https://go.dev/doc/", Acessos: 7,
		Criado: time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)}
	casos := casosFalsos{
		resolver:  func(context.Context, string) (dominio.Link, error) { return link, nil },
		consultar: func(context.Context, string) (dominio.Link, error) { return link, nil },
	}

	rec := chamar(casos, "GET", "/abc234", "")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != link.Destino {
		t.Errorf("redirecionar: %d %q", rec.Code, rec.Header().Get("Location"))
	}

	rec = chamar(casos, "GET", "/links/abc234", "")
	var dto contrato.Link
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
		t.Fatal(err)
	}
	if dto.Acessos != 7 || dto.Destino != link.Destino || !dto.Criado.Equal(link.Criado) {
		t.Errorf("dto = %+v", dto)
	}
}

func TestCorpoInvalido(t *testing.T) {
	naoDeveSerChamado := func(context.Context, string) (dominio.Link, error) {
		t.Error("o caso de uso foi chamado com um corpo inválido")
		return dominio.Link{}, nil
	}
	casos := casosFalsos{criar: naoDeveSerChamado}
	for nome, corpo := range map[string]string{
		"json quebrado": `{`,
		"corpo enorme":  `{"url":"http://x.org/` + strings.Repeat("a", 5000) + `"}`,
	} {
		t.Run(nome, func(t *testing.T) {
			if rec := chamar(casos, "POST", "/links", corpo); rec.Code != http.StatusBadRequest {
				t.Errorf("status %d", rec.Code)
			}
		})
	}
}

func TestSaudeEMetodoErrado(t *testing.T) {
	if rec := chamar(casosFalsos{}, "GET", "/saude", ""); rec.Code != 200 {
		t.Errorf("saúde: %d", rec.Code)
	}
	if rec := chamar(casosFalsos{}, "DELETE", "/links", ""); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE: %d", rec.Code)
	}
}
