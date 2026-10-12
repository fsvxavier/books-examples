// Package inl mostra o orçamento de inlining do compilador.
package inl

// Soma é pequena: cabe no orçamento de 80 nós.
func Soma(a, b int) int { return a + b }

// Limita é um pouco maior, mas ainda cabe.
func Limita(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Descricao tem laço, fmt-like concatenação e vários ramos: passa do orçamento.
func Descricao(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		switch {
		case i%15 == 0:
			s += "fizzbuzz,"
		case i%5 == 0:
			s += "buzz,"
		case i%3 == 0:
			s += "fizz,"
		default:
			s += "n,"
		}
	}
	return s
}

// Usa chama as três para vermos quem é expandida no chamador.
func Usa(n int) (int, int, int) {
	return Soma(n, 1), Limita(n, 0, 10), len(Descricao(n))
}

//go:noinline
func Fixa(a, b int) int { return a + b }

// UsaFixa chama a função marcada com //go:noinline.
func UsaFixa(n int) int { return Fixa(n, 1) }
