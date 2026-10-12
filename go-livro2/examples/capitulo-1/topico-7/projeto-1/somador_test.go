package e07

import "testing"

type Somador interface{ Somar(a, b int) int }

type soma struct{}

func (soma) Somar(a, b int) int { return a + b }

//go:noinline
func viaInterface(s Somador, n int) int {
	t := 0
	for i := range n {
		t = s.Somar(t, i)
	}
	return t
}

//go:noinline
func viaConcreto(s soma, n int) int {
	t := 0
	for i := range n {
		t = s.Somar(t, i)
	}
	return t
}

//go:noinline
func viaFuncao(f func(a, b int) int, n int) int {
	t := 0
	for i := range n {
		t = f(t, i)
	}
	return t
}

//go:noinline
func viaGenerico[S Somador](s S, n int) int {
	t := 0
	for i := range n {
		t = s.Somar(t, i)
	}
	return t
}

var sink int

const N = 1000

func BenchmarkConcreto(b *testing.B) {
	for b.Loop() {
		sink = viaConcreto(soma{}, N)
	}
}

func BenchmarkInterface(b *testing.B) {
	var s Somador = soma{}
	for b.Loop() {
		sink = viaInterface(s, N)
	}
}

func BenchmarkFuncao(b *testing.B) {
	f := func(a, b int) int { return a + b }
	for b.Loop() {
		sink = viaFuncao(f, N)
	}
}

func BenchmarkGenerico(b *testing.B) {
	for b.Loop() {
		sink = viaGenerico(soma{}, N)
	}
}
