package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

// main sobe o handler em um servidor de teste local e faz duas chamadas: uma
// dentro do prazo e outra que excede o prazo (a resposta é 504).
func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	for _, c := range []struct {
		nome            string
		prazo, trabalho time.Duration
	}{
		{"dentro do prazo", 500 * time.Millisecond, 10 * time.Millisecond},
		{"excede o prazo", 20 * time.Millisecond, 2 * time.Second},
	} {
		srv := httptest.NewServer(ComTraceID(geradorSequencial(), NovoHandler(c.prazo, c.trabalho, log)))
		resp, err := http.Get(srv.URL + "/pedido")
		if err != nil {
			fmt.Println(c.nome, "erro:", err)
			srv.Close()
			continue
		}
		corpo, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("%-16s status=%d trace=%s corpo=%q\n", c.nome, resp.StatusCode, resp.Header.Get("X-Trace-ID"), corpo)
		srv.Close()
	}
}
