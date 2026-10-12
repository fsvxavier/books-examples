// Command e06 mostra os métodos genéricos do Go 1.27.
package main

import (
	"fmt"
	"strconv"
)

// Lista é uma lista de valores de T.
type Lista[T any] struct{ itens []T }

// De monta uma lista.
func De[T any](itens ...T) Lista[T] { return Lista[T]{itens: itens} }

// Filtrar mantém os itens para os quais f é verdadeira. Não precisa de outro tipo.
func (l Lista[T]) Filtrar(f func(T) bool) Lista[T] {
	var s Lista[T]
	for _, v := range l.itens {
		if f(v) {
			s.itens = append(s.itens, v)
		}
	}
	return s
}

// Mapear transforma cada item em um R: o método declara o seu PRÓPRIO parâmetro de tipo (Go 1.27).
func (l Lista[T]) Mapear[R any](f func(T) R) Lista[R] {
	s := Lista[R]{itens: make([]R, 0, len(l.itens))}
	for _, v := range l.itens {
		s.itens = append(s.itens, f(v))
	}
	return s
}

// Reduzir junta os itens em um valor de tipo A.
func (l Lista[T]) Reduzir[A any](inicial A, f func(A, T) A) A {
	acc := inicial
	for _, v := range l.itens {
		acc = f(acc, v)
	}
	return acc
}

// Itens devolve os valores.
func (l Lista[T]) Itens() []T { return l.itens }

// Caixa guarda um valor e sabe convertê-lo.
type Caixa struct{ valor int }

// Como devolve o valor da caixa no tipo pedido, a partir de uma função de conversão.
func (c Caixa) Como[R any](converter func(int) R) R { return converter(c.valor) }

func main() {
	numeros := De(1, 2, 3, 4, 5, 6)

	// Encadeamento: o que antes exigia funções soltas (Mapear(Filtrar(...))) lê-se da esquerda para a direita.
	textos := numeros.
		Filtrar(func(n int) bool { return n%2 == 0 }).
		Mapear(func(n int) string { return "#" + strconv.Itoa(n) })
	fmt.Printf("%q\n", textos.Itens())

	soma := numeros.Reduzir(0.5, func(acc float64, n int) float64 { return acc + float64(n) })
	fmt.Println("Reduzir:", soma)

	c := Caixa{valor: 7}
	fmt.Println(c.Como(strconv.Itoa), c.Como(func(n int) bool { return n > 5 }))

	// Instanciação explícita no método, e valor de método com o tipo fixado.
	m := numeros.Mapear[float64]
	fmt.Println(m(func(n int) float64 { return float64(n) / 2 }).Itens())
	fmt.Printf("tipo do valor de método: %T\n", numeros.Mapear[string])
}
