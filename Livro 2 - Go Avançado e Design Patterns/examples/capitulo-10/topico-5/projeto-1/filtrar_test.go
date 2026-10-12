package main

import (
	"slices"
	"testing"
)

func TestLimitarPara(t *testing.T) {
	var n int
	got := slices.Collect(Limitar(Contar(Naturais(), &n), 3))
	if !slices.Equal(got, []int{1, 2, 3}) || n != 3 {
		t.Fatalf("got=%v origem=%d: a origem nao deveria passar do 3o", got, n)
	}
}

func TestLimitarZeroNaoConsomeNada(t *testing.T) {
	var n int
	_ = slices.Collect(Limitar(Contar(Naturais(), &n), 0))
	if n != 0 {
		t.Fatalf("origem produziu %d", n)
	}
}

func TestPreguica(t *testing.T) {
	var c int
	s := Mapear(slices.Values([]int{1, 2, 3}), func(n int) int { c++; return n })
	if c != 0 {
		t.Fatal("o pipeline avaliou antes de ser consumido")
	}
	_ = slices.Collect(s)
	_ = slices.Collect(s) // uma Seq pode ser consumida de novo: e uma receita
	if c != 6 {
		t.Fatalf("c = %d", c)
	}
}

func TestZipTerminaComAMenor(t *testing.T) {
	var pares []string
	for a, b := range Zip(slices.Values([]string{"x", "y"}), Naturais()) {
		pares = append(pares, a+string(rune('0'+b)))
	}
	if !slices.Equal(pares, []string{"x1", "y2"}) {
		t.Fatalf("pares = %v", pares)
	}
}

func TestZipComBreakTermina(t *testing.T) {
	for range 1000 {
		for range Zip(Naturais(), Naturais()) {
			break
		}
	}
}

func TestFiltrarEMapear(t *testing.T) {
	got := slices.Collect(Mapear(Filtrar(slices.Values([]int{1, 2, 3, 4}), func(n int) bool { return n%2 == 0 }), func(n int) string { return string(rune('a' + n)) }))
	if !slices.Equal(got, []string{"c", "e"}) {
		t.Fatalf("got = %v", got)
	}
}
