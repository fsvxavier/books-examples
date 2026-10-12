package main

import (
	"fmt"
	"net/http"
	"sync"
)

// Middleware e um decorador de http.Handler: recebe um Handler e devolve outro.
type Middleware func(http.Handler) http.Handler

// Encadear aplica os middlewares de modo que o PRIMEIRO da lista fique mais por FORA
// (e portanto rode primeiro na ida e por ultimo na volta).
func Encadear(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Diario guarda, em ordem, o que aconteceu (seguro para uso concorrente).
type Diario struct {
	mu     sync.Mutex
	Linhas []string
}

func (d *Diario) Add(f string, a ...any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Linhas = append(d.Linhas, fmt.Sprintf(f, a...))
}

// Rastro registra a entrada e a saida: mostra a ordem dos decoradores.
func Rastro(nome string, d *Diario) Middleware {
	return func(prox http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			d.Add("-> %s", nome)
			prox.ServeHTTP(w, r)
			d.Add("<- %s", nome)
		})
	}
}

// gravador captura o status que o handler interno escreveu.
type gravador struct {
	http.ResponseWriter
	status int
}

func (g *gravador) WriteHeader(c int) { g.status = c; g.ResponseWriter.WriteHeader(c) }

func (g *gravador) Unwrap() http.ResponseWriter { return g.ResponseWriter }

// Registrar anota metodo, caminho e status DEPOIS da resposta.
func Registrar(d *Diario) Middleware {
	return func(prox http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			g := &gravador{ResponseWriter: w, status: http.StatusOK}
			prox.ServeHTTP(g, r)
			d.Add("log: %s %s -> %d", r.Method, r.URL.Path, g.status)
		})
	}
}

// Autenticar recusa requisicoes sem o token certo (e NAO chama o proximo: curto-circuito).
func Autenticar(token string) Middleware {
	return func(prox http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "nao autorizado", http.StatusUnauthorized)
				return
			}
			prox.ServeHTTP(w, r)
		})
	}
}

// Recuperar transforma panic em 500, em vez de derrubar a conexao.
func Recuperar(d *Diario) Middleware {
	return func(prox http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if e := recover(); e != nil {
					d.Add("panic recuperado: %v", e)
					http.Error(w, "erro interno", http.StatusInternalServerError)
				}
			}()
			prox.ServeHTTP(w, r)
		})
	}
}
