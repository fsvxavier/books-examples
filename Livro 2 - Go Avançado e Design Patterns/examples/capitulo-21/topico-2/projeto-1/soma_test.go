package soma

import "testing"

// Estilo antigo: o resultado nao e usado, e o compilador pode apagar a chamada.
func BenchmarkSomaAntigo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Soma(1, 2)
	}
}

// Estilo atual (Go 1.24): b.Loop mantem a chamada viva.
func BenchmarkSomaLoop(b *testing.B) {
	for b.Loop() {
		Soma(1, 2)
	}
}

var sink int

// Estilo antigo, do jeito certo: o resultado vai para uma variavel de pacote.
func BenchmarkSomaAntigoComSink(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Soma(1, 2)
	}
}

func BenchmarkFib20(b *testing.B) {
	for b.Loop() {
		Fib(20)
	}
}
