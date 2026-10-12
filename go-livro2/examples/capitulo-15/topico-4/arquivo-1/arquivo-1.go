package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
)

// ---- Versão ruim: cada handler repete a verificação. Um handler novo esqueceu. ----

func pedidosRuim(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") == "" {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	fmt.Fprintln(w, "lista de pedidos")
}

func relatorioRuim(w http.ResponseWriter, r *http.Request) {
	// novo endpoint, escrito às pressas: ninguém lembrou da autenticação
	fmt.Fprintln(w, "relatório financeiro confidencial")
}

// ---- Versão boa: o comportamento transversal vive em middlewares. ----

type Middleware func(http.Handler) http.Handler

func Encadear(h http.Handler, mws ...Middleware) http.Handler {
	for _, m := range slices.Backward(mws) {
		h = m(h)
	}
	return h
}

func Autenticacao(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "não autenticado", http.StatusUnauthorized)
			return
		}
		proximo.ServeHTTP(w, r)
	})
}

// registrador captura o status escrito pelo handler.
type registrador struct {
	http.ResponseWriter
	status int
}

func (g *registrador) WriteHeader(code int) {
	g.status = code
	g.ResponseWriter.WriteHeader(code)
}

func ComLog(logger *slog.Logger) Middleware {
	return func(proximo http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			g := &registrador{ResponseWriter: w, status: http.StatusOK}
			proximo.ServeHTTP(g, r)
			logger.Info("req", "metodo", r.Method, "path", r.URL.Path, "status", g.status)
		})
	}
}

func logSemTempo() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
}

func chamar(h http.Handler, caminho, auth string) {
	req := httptest.NewRequest("GET", caminho, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	corpo := strings.TrimSpace(rec.Body.String())
	fmt.Printf("  %-12s auth=%-5q -> %d %s\n", caminho, auth, rec.Code, corpo)
}

func main() {
	fmt.Println("versão ruim (cada handler cuida de si):")
	mux := http.NewServeMux()
	mux.HandleFunc("/pedidos", pedidosRuim)
	mux.HandleFunc("/relatorio", relatorioRuim)
	chamar(mux, "/pedidos", "")
	chamar(mux, "/relatorio", "")

	fmt.Println("versão boa (um ponto só):")
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/pedidos", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "lista de pedidos\n") })
	mux2.HandleFunc("/relatorio", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "relatório financeiro confidencial\n")
	})
	h := Encadear(mux2, ComLog(logSemTempo()), Autenticacao)
	chamar(h, "/pedidos", "")
	chamar(h, "/relatorio", "")
	chamar(h, "/relatorio", "token")

	fmt.Println("ordem invertida (Autenticacao por fora do log):")
	h2 := Encadear(mux2, Autenticacao, ComLog(logSemTempo()))
	chamar(h2, "/relatorio", "")
	chamar(h2, "/relatorio", "token")
}
