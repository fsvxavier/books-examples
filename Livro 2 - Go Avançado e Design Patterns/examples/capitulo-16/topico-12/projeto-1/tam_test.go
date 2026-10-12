package e10

import (
	"math/rand/v2"
	"slices"
	"sort"
	"testing"
)

const tam = 100_000

func dados() []int {
	r := rand.New(rand.NewPCG(1, 2))
	s := make([]int, tam)
	for i := range s {
		s[i] = r.IntN(1 << 30)
	}
	return s
}

func BenchmarkOrdenar_sortInts(b *testing.B) {
	base := dados()
	b.ReportAllocs()
	for b.Loop() {
		s := slices.Clone(base)
		sort.Ints(s)
	}
}

func BenchmarkOrdenar_sortSlice(b *testing.B) {
	base := dados()
	b.ReportAllocs()
	for b.Loop() {
		s := slices.Clone(base)
		sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	}
}

func BenchmarkOrdenar_slicesSort(b *testing.B) {
	base := dados()
	b.ReportAllocs()
	for b.Loop() {
		s := slices.Clone(base)
		slices.Sort(s)
	}
}

func BenchmarkOrdenar_slicesSortFunc(b *testing.B) {
	base := dados()
	b.ReportAllocs()
	for b.Loop() {
		s := slices.Clone(base)
		slices.SortFunc(s, func(a, b int) int { return a - b })
	}
}

// Somar com qualquer coisa que tenha Valor(): interface (despacho dinâmico) x constraint genérica.
type valorer interface{ Valor() int }

type inteiro int

func (i inteiro) Valor() int { return int(i) }

func somaInterface(xs []valorer) int {
	t := 0
	for _, x := range xs {
		t += x.Valor()
	}
	return t
}

func somaGenerica[T valorer](xs []T) int {
	t := 0
	for _, x := range xs {
		t += x.Valor()
	}
	return t
}

func somaDireta(xs []inteiro) int {
	t := 0
	for _, x := range xs {
		t += int(x)
	}
	return t
}

var sink int

func BenchmarkSoma_interface(b *testing.B) {
	xs := make([]valorer, tam)
	for i := range xs {
		xs[i] = inteiro(i)
	}
	for b.Loop() {
		sink = somaInterface(xs)
	}
}

func BenchmarkSoma_generica(b *testing.B) {
	xs := make([]inteiro, tam)
	for i := range xs {
		xs[i] = inteiro(i)
	}
	for b.Loop() {
		sink = somaGenerica(xs)
	}
}

func BenchmarkSoma_generica_viaInterface(b *testing.B) {
	xs := make([]valorer, tam)
	for i := range xs {
		xs[i] = inteiro(i)
	}
	for b.Loop() {
		sink = somaGenerica(xs) // T = valorer: o despacho continua dinâmico
	}
}

func BenchmarkSoma_direta(b *testing.B) {
	xs := make([]inteiro, tam)
	for i := range xs {
		xs[i] = inteiro(i)
	}
	for b.Loop() {
		sink = somaDireta(xs)
	}
}
