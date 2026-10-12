package main

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type espiao struct {
	base
	chamado *int
}

func (e *espiao) Tratar(ctx context.Context, r *Requisicao) error {
	*e.chamado++
	return e.repassar(ctx, r)
}

func TestFimDaCadeiaPassaSemErro(t *testing.T) {
	var n int
	e := &espiao{chamado: &n}
	if err := e.Tratar(context.Background(), &Requisicao{}); err != nil || n != 1 {
		t.Fatalf("err=%v n=%d", err, n)
	}
}

func TestAutenticacaoInterrompeACadeia(t *testing.T) {
	var depois int
	a := &Autenticacao{Usuarios: map[string]string{"ana": "1"}}
	a.DefinirProximo(&espiao{chamado: &depois})
	err := a.Tratar(context.Background(), &Requisicao{Usuario: "ana", Senha: "errada"})
	if !errors.Is(err, ErrNaoAutenticado) || depois != 0 {
		t.Fatalf("err=%v depois=%d: o proximo elo nao deveria ter rodado", err, depois)
	}
}

func TestAutenticacaoMarcaERepassa(t *testing.T) {
	var depois int
	a := &Autenticacao{Usuarios: map[string]string{"ana": "1"}}
	a.DefinirProximo(&espiao{chamado: &depois})
	r := &Requisicao{Usuario: "ana", Senha: "1"}
	if err := a.Tratar(context.Background(), r); err != nil || !r.Autenticado || depois != 1 {
		t.Fatalf("err=%v aut=%v depois=%d", err, r.Autenticado, depois)
	}
}

func TestCacheEncerraComSucesso(t *testing.T) {
	var depois int
	c := NovoCache()
	c.DefinirProximo(&espiao{chamado: &depois})
	r := &Requisicao{Corpo: "x"}
	c.Tratar(context.Background(), r) // 1a vez: passa adiante e guarda
	c.Tratar(context.Background(), r) // 2a vez: o cache responde
	if depois != 1 || c.Acertos != 1 {
		t.Fatalf("depois=%d acertos=%d", depois, c.Acertos)
	}
}

func TestCacheNaoGuardaFalha(t *testing.T) {
	c := NovoCache()
	c.DefinirProximo(&Sanitizacao{})
	for range 3 {
		if err := c.Tratar(context.Background(), &Requisicao{Corpo: "<x>"}); !errors.Is(err, ErrInvalida) {
			t.Fatalf("err = %v", err)
		}
	}
	if c.Acertos != 0 {
		t.Fatal("uma requisicao que falhou foi servida do cache")
	}
}

func TestContextoCanceladoNoFimDaCadeia(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &Pedidos{}
	if err := p.Tratar(ctx, &Requisicao{}); !errors.Is(err, context.Canceled) || len(p.Criados) != 0 {
		t.Fatalf("err=%v criados=%v", err, p.Criados)
	}
}

func TestOrdemMuda(t *testing.T) {
	novaReq := func() *Requisicao { return &Requisicao{Usuario: "ana", Senha: "x", IP: "ip", Corpo: "c"} }
	usuarios := map[string]string{"ana": "1"}

	// limite POR DENTRO da autenticacao: nunca ve as falhas.
	a1 := &Autenticacao{Usuarios: usuarios}
	a1.DefinirProximo(NovoLimite(1))
	for range 5 {
		if err := a1.Tratar(context.Background(), novaReq()); errors.Is(err, ErrBloqueado) {
			t.Fatal("o limite por dentro nao deveria bloquear")
		}
	}

	// limite POR FORA: bloqueia depois de 1 falha.
	l := NovoLimite(1)
	l.DefinirProximo(&Autenticacao{Usuarios: usuarios})
	l.Tratar(context.Background(), novaReq())
	if err := l.Tratar(context.Background(), novaReq()); !errors.Is(err, ErrBloqueado) {
		t.Fatalf("err = %v", err)
	}
}

func TestCadeiaCompletaConcorrente(t *testing.T) {
	auth := &Autenticacao{Usuarios: map[string]string{"ana": "1"}}
	pedidos := &Pedidos{}
	auth.DefinirProximo(NovoLimite(1000)).DefinirProximo(&Sanitizacao{}).DefinirProximo(NovoCache()).DefinirProximo(pedidos)
	var wg sync.WaitGroup
	for i := range 200 {
		wg.Go(func() {
			r := &Requisicao{Usuario: "ana", Senha: "1", IP: "ip", Corpo: string(rune('a' + i%10))}
			if err := auth.Tratar(context.Background(), r); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if len(pedidos.Criados) < 10 {
		t.Fatalf("criados = %d", len(pedidos.Criados))
	}
}
