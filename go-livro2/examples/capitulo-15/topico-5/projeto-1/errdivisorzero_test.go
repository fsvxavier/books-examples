package e04

import "testing"

var sink int

func BenchmarkErroValor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := DividirErro(10, i%2) // metade das chamadas falha
		if err != nil {
			sink++
		}
	}
}

func BenchmarkPanicRecover(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := DividirPanicSeguro(10, i%2)
		if err != nil {
			sink++
		}
	}
}

func BenchmarkPanicRecoverSoSucesso(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DividirPanicSeguro(10, 1)
	}
}

func TestAnalisar(t *testing.T) {
	for _, s := range []string{"[1,22,333]", "[1,,3]", "[1,2", "[x]"} {
		func() {
			defer func() {
				if p := recover(); p != nil {
					t.Logf("%-12q pânico não tratado: %v", s, p)
				}
			}()
			v, err := Analisar(s)
			t.Logf("%-12q v=%v err=%v", s, v, err)
		}()
	}
}
