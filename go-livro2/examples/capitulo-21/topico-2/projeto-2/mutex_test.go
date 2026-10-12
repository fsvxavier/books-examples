package cont

import (
	"strconv"
	"strings"
	"testing"
)

func BenchmarkContador(b *testing.B) {
	b.Run("mutex", func(b *testing.B) {
		var c Mutex
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				c.Inc()
			}
		})
	})
	b.Run("atomico", func(b *testing.B) {
		var c Atomico
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				c.Inc()
			}
		})
	})
}

// Montar 1.000 numeros em texto, e reportar o volume processado.
func BenchmarkJuntar(b *testing.B) {
	partes := make([]string, 1000)
	total := 0
	for i := range partes {
		partes[i] = strconv.Itoa(i)
		total += len(partes[i])
	}
	b.SetBytes(int64(total))
	for b.Loop() {
		_ = strings.Join(partes, ",")
	}
	b.ReportMetric(float64(len(partes)), "itens/op")
}
