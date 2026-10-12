package crc

import (
	"hash/crc32"
	"math/rand/v2"
	"testing"
)

func TestSomarIgualAoReferencia(t *testing.T) {
	t.Log("implementacao:", Implementacao)
	r := rand.New(rand.NewPCG(1, 2))
	for _, n := range []int{0, 1, 7, 64, 1000, 1 << 20} {
		dados := make([]byte, n)
		for i := range dados {
			dados[i] = byte(r.IntN(256))
		}
		quer := crc32.ChecksumIEEE(dados)
		if got := Somar(0, dados); got != quer {
			t.Errorf("n=%d: Somar = %#x, quer %#x", n, got, quer)
		}
	}
}

func TestValorConhecido(t *testing.T) {
	// valor de referencia publicado para o CRC-32 de "123456789"
	if got := Somar(0, []byte("123456789")); got != 0xcbf43926 {
		t.Errorf("Somar = %#x, quer 0xcbf43926", got)
	}
}

func TestSomarEmPedacos(t *testing.T) {
	dados := []byte("a pressa e inimiga da perfeicao")
	inteiro := Somar(0, dados)
	parcial := Somar(Somar(0, dados[:10]), dados[10:])
	if inteiro != parcial {
		t.Errorf("em pedacos: %#x, inteiro: %#x", parcial, inteiro)
	}
}

func BenchmarkSomar(b *testing.B) {
	for _, n := range []int{16, 256, 4 << 10, 1 << 20} {
		dados := make([]byte, n)
		b.Run(tamanho(n), func(b *testing.B) {
			b.SetBytes(int64(n))
			for b.Loop() {
				Somar(0, dados)
			}
		})
	}
}

func tamanho(n int) string {
	switch {
	case n >= 1<<20:
		return "1MiB"
	case n >= 1<<10:
		return "4KiB"
	}
	return map[int]string{16: "16B", 256: "256B"}[n]
}
