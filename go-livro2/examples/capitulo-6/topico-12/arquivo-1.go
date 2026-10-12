package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

// semUnwrap embrulha o ResponseWriter mas SO expoe os metodos da interface basica.
type semUnwrap struct{ http.ResponseWriter }

// comUnwrap faz o mesmo e informa quem ele embrulha.
type comUnwrap struct{ http.ResponseWriter }

func (c comUnwrap) Unwrap() http.ResponseWriter { return c.ResponseWriter }

func main() {
	handler := func(embrulhar func(http.ResponseWriter) http.ResponseWriter) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w = embrulhar(w)
			fmt.Fprint(w, "parte 1;")
			// Flush envia o que ja foi escrito sem esperar o fim (streaming, SSE).
			err := http.NewResponseController(w).Flush()
			if err != nil {
				fmt.Fprintf(w, " erro no Flush: %v", err)
				return
			}
			fmt.Fprint(w, " parte 2")
		})
	}

	for _, c := range []struct {
		nome string
		emb  func(http.ResponseWriter) http.ResponseWriter
	}{
		{"sem embrulho", func(w http.ResponseWriter) http.ResponseWriter { return w }},
		{"embrulho SEM Unwrap", func(w http.ResponseWriter) http.ResponseWriter { return semUnwrap{w} }},
		{"embrulho COM Unwrap", func(w http.ResponseWriter) http.ResponseWriter { return comUnwrap{w} }},
	} {
		srv := httptest.NewServer(handler(c.emb))
		resp, _ := http.Get(srv.URL)
		corpo, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		srv.Close()
		fmt.Printf("%-20s -> %s\n", c.nome, corpo)
	}

	// Assercao de tipo direta: o embrulho esconde as interfaces opcionais.
	rec := httptest.NewRecorder()
	_, direto := http.ResponseWriter(rec).(http.Flusher)
	_, embrulhado := http.ResponseWriter(semUnwrap{rec}).(http.Flusher)
	fmt.Println("recorder e Flusher:", direto, "| embrulhado em semUnwrap e Flusher:", embrulhado)
}
