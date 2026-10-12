package desconto

import (
	"fmt"
	"math"
	"testing"
)

// Teste de caracterização: fixa o que o código ANTIGO faz hoje, inclusive nas bordas.
func TestCaracterizacao(t *testing.T) {
	casos := []struct {
		valor float64
		want  float64
	}{
		{0, 0},
		{100, 5},
		{1000, 50},         // 1000 exato NÃO é premium
		{1000.01, 100.001}, // um centavo acima já é premium: o desconto "salta"
	}
	for _, c := range casos {
		if got := Antigo(c.valor); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Antigo(%v) = %v, quer %v", c.valor, got, c.want)
		}
	}
}

// Teste diferencial: compara antigo e novo em todos os valores de 0,00 a 3000,00.
func diferencial(novo func(int64) int64) (diverge, meioCentavo int, exemplos []string) {
	for c := int64(0); c <= 300_000; c++ {
		antigo := Antigo(float64(c) / 100)
		want := int64(math.Round(antigo * 100))
		if got := novo(c); got != want {
			diverge++
			pct := int64(percentualPadrao)
			if c > limiarPremium {
				pct = percentualPremium
			}
			if c*pct%100 == 50 { // o desconto exato cai no meio de um centavo
				meioCentavo++
			}
			if len(exemplos) < 3 {
				exemplos = append(exemplos, fmt.Sprintf("R$ %d,%02d: antigo=%.5f -> %d centavos, novo=%d", c/100, c%100, antigo, want, got))
			}
		}
	}
	return
}

func TestDiferencial(t *testing.T) {
	for _, v := range []struct {
		nome string
		f    func(int64) int64
	}{{"NovoTruncado", NovoTruncado}, {"NovoArredondado", NovoArredondado}} {
		n, meio, ex := diferencial(v.f)
		t.Logf("%-16s divergências em 300001 valores: %d (das quais no meio exato de um centavo: %d)", v.nome, n, meio)
		for _, e := range ex {
			t.Log("    ", e)
		}
	}
}
