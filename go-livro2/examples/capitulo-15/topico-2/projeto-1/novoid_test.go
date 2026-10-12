package e01

import (
	"sync"
	"testing"
)

func TestGlobalMudaMoeda(t *testing.T) {
	config["moeda"] = "USD" // este teste "suja" o estado do processo inteiro
	if Moeda() != "USD" {
		t.Fatal("moeda deveria ser USD")
	}
}

func TestGlobalMoedaPadrao(t *testing.T) {
	if Moeda() != "BRL" {
		t.Fatalf("esperava BRL, veio %s", Moeda())
	}
}

func TestGlobalIDsConcorrentes(t *testing.T) {
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() { NovoID() })
	}
	wg.Wait()
}

func TestInstanciaIsolada(t *testing.T) {
	t.Parallel()
	loja := NovaLoja("USD", &Gerador{})
	if loja.Moeda != "USD" {
		t.Fatal("moeda")
	}
}

func TestInstanciaOutra(t *testing.T) {
	t.Parallel()
	loja := NovaLoja("BRL", &Gerador{})
	if loja.Moeda != "BRL" {
		t.Fatal("moeda")
	}
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() { loja.IDs.NovoID() })
	}
	wg.Wait()
	if loja.IDs.NovoID() != 51 {
		t.Fatal("id")
	}
}
