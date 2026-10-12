package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"time"
)

func get(c *http.Client, url, token string) (int, string) {
	req, _ := http.NewRequest("GET", url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, strings.TrimSpace(string(b))
}

func main() {
	app := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/explode" {
			panic("bug no handler")
		}
		fmt.Fprint(w, "ola")
	})

	fmt.Println("== Ordem 1: Registrar(fora), Autenticar(dentro): negados TAMBEM sao registrados")
	d1 := &Diario{}
	s1 := httptest.NewServer(Encadear(app, Registrar(d1), Autenticar("segredo")))
	get(s1.Client(), s1.URL+"/", "")
	get(s1.Client(), s1.URL+"/", "segredo")
	s1.Close()
	for _, l := range d1.Linhas {
		fmt.Println("  ", l)
	}

	fmt.Println("== Ordem 2: Autenticar(fora), Registrar(dentro): negados NAO chegam ao registro")
	d2 := &Diario{}
	s2 := httptest.NewServer(Encadear(app, Autenticar("segredo"), Registrar(d2)))
	get(s2.Client(), s2.URL+"/", "")
	get(s2.Client(), s2.URL+"/", "segredo")
	s2.Close()
	for _, l := range d2.Linhas {
		fmt.Println("  ", l)
	}

	fmt.Println("== Entrada e saida: o primeiro da lista e o mais externo")
	d3 := &Diario{}
	s3 := httptest.NewServer(Encadear(app, Rastro("A", d3), Rastro("B", d3), Rastro("C", d3)))
	get(s3.Client(), s3.URL+"/", "")
	s3.Close()
	fmt.Println("  ", strings.Join(d3.Linhas, " "))

	fmt.Println("== Recuperar converte panic em 500")
	d4 := &Diario{}
	s4 := httptest.NewServer(Encadear(app, Registrar(d4), Recuperar(d4)))
	st, corpo := get(s4.Client(), s4.URL+"/explode", "")
	s4.Close()
	fmt.Println("  ", st, corpo)
	for _, l := range d4.Linhas {
		fmt.Println("  ", l)
	}

	fmt.Println("== Decorador de CLIENTE (RoundTripper): cabecalho + retentativa em 503")
	var chamadas atomic.Int32
	var visto atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visto.Store(r.Header.Get("X-App"))
		if chamadas.Add(1) < 3 {
			http.Error(w, "indisponivel", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprint(w, "finalmente")
	}))
	defer srv.Close()
	var esperas []time.Duration
	cliente := &http.Client{Transport: Decorar(http.DefaultTransport,
		ComRetentativa(5, func(d time.Duration) { esperas = append(esperas, d) }),
		ComCabecalho("X-App", "livro"),
	)}
	st, corpo = get(cliente, srv.URL, "")
	fmt.Println("  ", st, corpo, "| chamadas ao servidor:", chamadas.Load(), "| esperas:", esperas, "| cabecalho visto:", visto.Load())
}
