package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

var instante = time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC)

// novoTeste monta um servidor com relógio e códigos previsíveis.
func novoTeste(t *testing.T, codigos ...string) *servidor {
	t.Helper()
	s := novoServidor(config{baseURL: "http://curto.test/"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.agora = func() time.Time { return instante }
	i := 0
	s.novoCodigo = func() string {
		c := codigos[min(i, len(codigos)-1)]
		i++
		return c
	}
	return s
}

func chamar(s *servidor, metodo, alvo, corpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, alvo, strings.NewReader(corpo))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	return rec
}

func TestCriarEConsultar(t *testing.T) {
	s := novoTeste(t, "abc234")

	rec := chamar(s, "POST", "/links", `{"url":"https://go.dev/doc/"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar: status %d, corpo %s", rec.Code, rec.Body)
	}
	var r respostaCriar
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if r.Curto != "http://curto.test/abc234" {
		t.Errorf("curto = %q", r.Curto)
	}

	rec = chamar(s, "GET", "/abc234", "")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://go.dev/doc/" {
		t.Fatalf("redirecionar: %d %q", rec.Code, rec.Header().Get("Location"))
	}

	rec = chamar(s, "GET", "/links/abc234", "")
	var l link
	if err := json.Unmarshal(rec.Body.Bytes(), &l); err != nil {
		t.Fatal(err)
	}
	if l.Acessos != 1 || !l.Criado.Equal(instante) {
		t.Errorf("link = %+v", l)
	}
}

func TestErros(t *testing.T) {
	casos := []struct {
		nome, metodo, alvo, corpo string
		status                    int
	}{
		{"json quebrado", "POST", "/links", `{`, 400},
		{"destino vazio", "POST", "/links", `{"url":""}`, 422},
		{"esquema errado", "POST", "/links", `{"url":"ftp://x.org/a"}`, 422},
		{"sem host", "POST", "/links", `{"url":"http:///a"}`, 422},
		{"corpo enorme", "POST", "/links", `{"url":"http://x.org/` + strings.Repeat("a", limiteCorpo) + `"}`, 400},
		{"código inexistente", "GET", "/zzzzzz", "", 404},
		{"consulta inexistente", "GET", "/links/zzzzzz", "", 404},
		{"método errado", "DELETE", "/links", "", 405},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			s := novoTeste(t, "abc234")
			if rec := chamar(s, c.metodo, c.alvo, c.corpo); rec.Code != c.status {
				t.Errorf("status %d, esperado %d, corpo %s", rec.Code, c.status, rec.Body)
			}
		})
	}
}

func TestColisaoTentaDeNovo(t *testing.T) {
	s := novoTeste(t, "aaaaaa", "aaaaaa", "bbbbbb")
	chamar(s, "POST", "/links", `{"url":"https://a.test"}`)
	rec := chamar(s, "POST", "/links", `{"url":"https://b.test"}`)
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), "bbbbbb") {
		t.Fatalf("esperava bbbbbb após a colisão: %d %s", rec.Code, rec.Body)
	}
}

func TestSemCodigoLivre(t *testing.T) {
	s := novoTeste(t, "aaaaaa")
	chamar(s, "POST", "/links", `{"url":"https://a.test"}`)
	rec := chamar(s, "POST", "/links", `{"url":"https://b.test"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAcessosConcorrentes(t *testing.T) {
	s := novoTeste(t, "abc234")
	chamar(s, "POST", "/links", `{"url":"https://go.dev"}`)

	var wg sync.WaitGroup
	for range 200 {
		wg.Go(func() { chamar(s, "GET", "/abc234", "") })
	}
	wg.Wait()

	l, err := s.armazem.buscar("abc234")
	if err != nil || l.Acessos != 200 {
		t.Fatalf("acessos = %d (%v), esperado 200", l.Acessos, err)
	}
}
