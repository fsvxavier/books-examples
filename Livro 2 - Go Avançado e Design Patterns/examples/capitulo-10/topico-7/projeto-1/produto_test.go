package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"slices"
	"testing"
)

func novo(t *testing.T) (*Cliente, *Servidor) {
	t.Helper()
	srv := &Servidor{}
	var ts *httptest.Server
	ts = httptest.NewServer(srv.Handler(func() string { return ts.URL }))
	t.Cleanup(ts.Close)
	return &Cliente{Base: ts.URL, HTTP: ts.Client()}, srv
}

func coletar(t *testing.T, c *Cliente, ctx context.Context) ([]int, error) {
	var ids []int
	for p, err := range c.Produtos(ctx) {
		if err != nil {
			return ids, err
		}
		ids = append(ids, p.ID)
	}
	return ids, nil
}

func TestTodasAsPaginas(t *testing.T) {
	c, srv := novo(t)
	ids, err := coletar(t, c, context.Background())
	if err != nil || len(ids) != 25 || ids[0] != 1 || ids[24] != 25 || srv.Requisicoes != 3 {
		t.Fatalf("ids=%d err=%v requisicoes=%d", len(ids), err, srv.Requisicoes)
	}
}

func TestBreakNaoBuscaMaisPaginas(t *testing.T) {
	c, srv := novo(t)
	for p, err := range c.Produtos(context.Background()) {
		if err != nil || p.ID == 3 {
			break
		}
	}
	if srv.Requisicoes != 1 {
		t.Fatalf("requisicoes = %d, quero 1", srv.Requisicoes)
	}
}

func TestBreakNaFronteiraDaPagina(t *testing.T) {
	c, srv := novo(t)
	for p := range c.Produtos(context.Background()) {
		if p.ID == 10 { // ultimo da pagina 1
			break
		}
	}
	if srv.Requisicoes != 1 {
		t.Fatalf("requisicoes = %d: buscou a pagina 2 sem precisar", srv.Requisicoes)
	}
}

func TestErroNoMeio(t *testing.T) {
	c, srv := novo(t)
	srv.FalharNaPagina = 2
	ids, err := coletar(t, c, context.Background())
	if err == nil || len(ids) != 10 {
		t.Fatalf("ids=%d err=%v: esperava 10 produtos e depois o erro", len(ids), err)
	}
}

func TestContextoCancelado(t *testing.T) {
	c, _ := novo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := coletar(t, c, ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

func TestCadaRangeFazNovasRequisicoes(t *testing.T) {
	c, srv := novo(t)
	seq := c.Produtos(context.Background())
	for range seq {
	}
	antes := srv.Requisicoes
	for range seq {
	}
	if srv.Requisicoes != 2*antes {
		t.Fatalf("antes=%d depois=%d: cada range deveria refazer as requisicoes", antes, srv.Requisicoes)
	}
}

func TestComSlicesCollect(t *testing.T) {
	c, _ := novo(t)
	var ids []int
	for p := range c.Produtos(context.Background()) {
		ids = append(ids, p.ID)
	}
	if !slices.IsSorted(ids) {
		t.Fatal("fora de ordem")
	}
}
