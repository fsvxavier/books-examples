package main

import (
	"context"
	"errors"
	"testing"
)

func TestPrimeiroQueTrataParaACadeia(t *testing.T) {
	var chamados []string
	reg := func(n string, trata bool) Tratador {
		return Novo(n, func(context.Context, *Chamado) (bool, error) { chamados = append(chamados, n); return trata, nil })
	}
	cd := Cadeia{reg("a", false), reg("b", true), reg("c", true)}
	c := &Chamado{}
	if err := cd.Tratar(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	if c.TratadoPor != "b" || len(chamados) != 2 {
		t.Fatalf("tratadoPor=%q chamados=%v", c.TratadoPor, chamados)
	}
}

func TestNaoTratado(t *testing.T) {
	cd, _ := Fabrica("robo")
	err := cd.Tratar(context.Background(), &Chamado{Assunto: "algo", Gravidade: 3})
	if !errors.Is(err, ErrNaoTratado) {
		t.Fatalf("err = %v", err)
	}
}

func TestErroDeUmTratadorInterrompe(t *testing.T) {
	boom := errors.New("boom")
	depois := false
	cd := Cadeia{
		Novo("ruim", func(context.Context, *Chamado) (bool, error) { return false, boom }),
		Novo("depois", func(context.Context, *Chamado) (bool, error) { depois = true; return true, nil }),
	}
	err := cd.Tratar(context.Background(), &Chamado{})
	if !errors.Is(err, boom) || depois {
		t.Fatalf("err=%v depois=%v", err, depois)
	}
}

func TestContextoCancelado(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cd, _ := Fabrica("atendente")
	if err := cd.Tratar(ctx, &Chamado{Gravidade: 1}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

func TestInserirNaoMutaAOriginal(t *testing.T) {
	cd, _ := Fabrica("robo", "atendente")
	nova := cd.Inserir(1, PorGravidade("x", 5))
	if len(cd) != 2 || len(nova) != 3 || nova[1].Nome() != "x" {
		t.Fatalf("cd=%d nova=%d", len(cd), len(nova))
	}
}

func TestFabricaRecusaNomeDesconhecido(t *testing.T) {
	if _, err := Fabrica("fantasma"); err == nil {
		t.Fatal("esperava erro")
	}
}

// Tabela: cada gravidade vai para o primeiro tratador capaz.
func TestEscalonamento(t *testing.T) {
	cd, _ := Fabrica("atendente", "especialista", "coordenacao")
	casos := map[int]string{1: "atendente", 2: "atendente", 3: "especialista", 4: "especialista", 5: "coordenacao"}
	for g, quero := range casos {
		c := &Chamado{Gravidade: g}
		if err := cd.Tratar(context.Background(), c); err != nil || c.TratadoPor != quero {
			t.Errorf("gravidade %d: tratadoPor=%q err=%v, quero %q", g, c.TratadoPor, err, quero)
		}
	}
}
