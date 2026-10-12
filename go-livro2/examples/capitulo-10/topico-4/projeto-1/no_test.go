package main

import (
	"slices"
	"testing"
)

func TestOrdens(t *testing.T) {
	a := exemplo()
	if got := slices.Collect(a.All()); !slices.Equal(got, []int{1, 2, 4, 5, 3, 6}) {
		t.Fatalf("profundidade = %v", got)
	}
	if got := slices.Collect(a.Largura()); !slices.Equal(got, []int{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("largura = %v", got)
	}
	if got := slices.Collect(a.Folhas()); !slices.Equal(got, []int{4, 5, 6}) {
		t.Fatalf("folhas = %v", got)
	}
}

func TestArvoreVazia(t *testing.T) {
	var a Arvore[int]
	if len(slices.Collect(a.All())) != 0 || len(slices.Collect(a.Largura())) != 0 || len(slices.Collect(a.Folhas())) != 0 {
		t.Fatal("arvore vazia deveria nao produzir nada")
	}
}

func TestBreakParaATravessia(t *testing.T) {
	a := exemplo()
	for v := range a.All() {
		if v == 2 {
			break
		}
	}
	if a.Visitas != 2 {
		t.Fatalf("visitas = %d: a travessia continuou depois do break", a.Visitas)
	}
}

func TestReusavel(t *testing.T) {
	a := exemplo()
	seq := a.All()
	if !slices.Equal(slices.Collect(seq), slices.Collect(seq)) {
		t.Fatal("a mesma Seq percorrida duas vezes deu resultados diferentes")
	}
}

func TestDoisPercursosIndependentes(t *testing.T) {
	a := exemplo()
	var x, y []int
	for v := range a.All() {
		x = append(x, v)
		for w := range a.Largura() { // percurso aninhado sobre a mesma arvore
			if w == 3 {
				break
			}
		}
	}
	y = slices.Collect(a.All())
	if !slices.Equal(x, y) {
		t.Fatalf("o percurso aninhado interferiu no externo: %v x %v", x, y)
	}
}

func TestSeq2(t *testing.T) {
	a := exemplo()
	var niveis []int
	for n := range a.ComNivel() {
		niveis = append(niveis, n)
	}
	if !slices.Equal(niveis, []int{0, 1, 2, 2, 1, 2}) {
		t.Fatalf("niveis = %v", niveis)
	}
}
