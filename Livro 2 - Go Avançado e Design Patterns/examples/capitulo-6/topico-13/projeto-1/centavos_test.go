package main

import (
	"sync"
	"testing"
)

var (
	caneta  = Produto{"Caneta", 1000}
	mochila = Produto{"Mochila", 20000}
)

func TestFolhaEContainerTratadosIgual(t *testing.T) {
	var a Item = caneta
	var b Item = NovoKit("k", caneta, mochila)
	if a.Preco() != 1000 || b.Preco() != 21000 {
		t.Fatalf("a=%v b=%v", a.Preco(), b.Preco())
	}
}

func TestAninhamentoProfundo(t *testing.T) {
	var it Item = caneta
	for range 50 {
		it = NovoKit("n", it, caneta)
	}
	if got := it.Preco(); got != 51*1000 {
		t.Fatalf("preco = %v", got)
	}
}

func TestOrdemDosDecoradores(t *testing.T) {
	base := NovoKit("k", caneta, mochila) // 21000
	a := Imposto(DescontoFixo(base, 1000), 1000).Preco()
	b := DescontoFixo(Imposto(base, 1000), 1000).Preco()
	if a != 22000 || b != 22100 {
		t.Fatalf("a=%d b=%d", a, b)
	}
}

func TestArredondamentoDoImposto(t *testing.T) {
	// 10,5% sobre R$ 0,10 = 1,05 centavo: arredonda para 1.
	if got := Imposto(Produto{"x", 10}, 1050).Preco(); got != 11 {
		t.Fatalf("got %d", got)
	}
	// 10,5% sobre R$ 0,20 = 2,1 centavos: arredonda para 2.
	if got := Imposto(Produto{"x", 20}, 1050).Preco(); got != 22 {
		t.Fatalf("got %d", got)
	}
}

func TestDescontoNaoNegativa(t *testing.T) {
	if got := DescontoFixo(caneta, 5000).Preco(); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestCacheCalculaUmaVezEInvalida(t *testing.T) {
	k := NovoKit("k", caneta)
	c := &ComCache{Item: k}
	c.Preco()
	c.Preco()
	if c.calculos.Load() != 1 {
		t.Fatalf("calculos = %d", c.calculos.Load())
	}
	k.Adicionar(mochila)
	if c.Preco() != 1000 {
		t.Fatal("o cache deveria estar velho")
	}
	c.Invalidar()
	if c.Preco() != 21000 {
		t.Fatal("depois de Invalidar deveria recalcular")
	}
}

func TestLeituraConcorrente(t *testing.T) {
	c := &ComCache{Item: NovoKit("k", caneta, mochila)}
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			if c.Preco() != 21000 {
				t.Error("preco errado")
			}
		})
	}
	wg.Wait()
}
