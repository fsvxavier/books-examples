package main

import (
	"context"
	"fmt"
	"net/http/httptest"
)

func main() {
	srv := &Servidor{}
	var ts *httptest.Server
	ts = httptest.NewServer(srv.Handler(func() string { return ts.URL }))
	defer ts.Close()
	c := &Cliente{Base: ts.URL, HTTP: ts.Client()}
	ctx := context.Background()

	// 1) Percorrer tudo: o chamador so ve um laco.
	n := 0
	for p, err := range c.Produtos(ctx) {
		if err != nil {
			fmt.Println("erro:", err)
			break
		}
		n++
		_ = p
	}
	fmt.Println("percorreu tudo:", n, "produtos com", srv.Requisicoes, "requisicoes")

	// 2) Parar cedo: nao busca as paginas seguintes.
	srv.Requisicoes = 0
	for p, err := range c.Produtos(ctx) {
		if err != nil {
			break
		}
		if p.ID == 7 {
			break
		}
	}
	fmt.Println("parou no produto 7:", srv.Requisicoes, "requisicao (so a pagina 1)")

	// 3) O erro chega pelo proprio laco, no ponto exato em que ocorreu.
	srv.Requisicoes, srv.FalharNaPagina = 0, 2
	n = 0
	for p, err := range c.Produtos(ctx) {
		if err != nil {
			fmt.Printf("erro depois de %d produtos: %v\n", n, err)
			break
		}
		n++
		_ = p
	}

	// 4) Contexto cancelado: o iterador respeita.
	srv.FalharNaPagina = 0
	cancelado, cancel := context.WithCancel(ctx)
	cancel()
	for _, err := range c.Produtos(cancelado) {
		fmt.Println("contexto cancelado:", err != nil)
		break
	}
}
