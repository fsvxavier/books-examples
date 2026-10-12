package escape

import "fmt"

type Ponto struct{ X, Y, Z float64 }

// PorPonteiro devolve o endereco de uma variavel local: ela precisa viver
// depois que a funcao retornar, entao vai para o heap.
//
//go:noinline
func PorPonteiro(x float64) *Ponto {
	p := Ponto{x, x, x}
	return &p
}

// PorValor devolve uma copia: a variavel fica na pilha.
//
//go:noinline
func PorValor(x float64) Ponto {
	return Ponto{x, x, x}
}

// Texto converte para interface (fmt.Sprint recebe any): o valor escapa.
//
//go:noinline
func Texto(n int) string {
	return fmt.Sprint(n)
}

// BufferFixo tem tamanho constante e nao escapa: fica na pilha.
//
//go:noinline
func BufferFixo() int {
	buf := make([]byte, 64)
	for i := range buf {
		buf[i] = byte(i)
	}
	return int(buf[10])
}

// BufferVariavel tem tamanho conhecido so em tempo de execucao: vai para o heap.
//
//go:noinline
func BufferVariavel(n int) int {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = byte(i)
	}
	return int(buf[10])
}

// Soma recebe um slice que so e lido: nada escapa.
func Soma(valores []int) int {
	t := 0
	for _, v := range valores {
		t += v
	}
	return t
}
