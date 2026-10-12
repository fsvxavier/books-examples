package e06

import (
	"fmt"
	"testing"
)

var sink string

func TestIguais(t *testing.T) {
	l := linhasDe(100)
	ref := Concat(l)
	for nome, f := range map[string]func([]string) string{"Builder": Builder, "Grow": BuilderGrow, "Buffer": Buffer, "Join": Join} {
		if f(l) != ref {
			t.Fatal(nome, "difere")
		}
	}
}

func BenchmarkConcat(b *testing.B) {
	for _, n := range []int{10, 1000, 100000} {
		l := linhasDe(n)
		fns := []struct {
			nome string
			f    func([]string) string
		}{{"Concat", Concat}, {"Builder", Builder}, {"BuilderGrow", BuilderGrow}, {"Buffer", Buffer}, {"Join", Join}}
		for _, x := range fns {
			b.Run(fmt.Sprintf("%s/n=%d", x.nome, n), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					sink = x.f(l)
				}
			})
		}
	}
}
