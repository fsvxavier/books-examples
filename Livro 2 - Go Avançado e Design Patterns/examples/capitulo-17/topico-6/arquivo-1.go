// Command e11 explora a inferência de tipos pelas constraints (constraint type inference).
package main

import "fmt"

type Numero interface{ ~int | ~float64 }

type Ids []int

// Soma infere E a partir de S: o core type de S é []E.
func Soma[S ~[]E, E Numero](s S) E {
	var t E
	for _, v := range s {
		t += v
	}
	return t
}

// Predicado é um tipo função genérico.
type Predicado[T any] func(T) bool

// Filtrar infere T da fatia, e o literal de função é convertido para Predicado[T].
func Filtrar[T any](xs []T, p Predicado[T]) []T {
	var s []T
	for _, x := range xs {
		if p(x) {
			s = append(s, x)
		}
	}
	return s
}

// Par é um tipo genérico.
type Par[A, B any] struct {
	Primeiro A
	Segundo  B
}

// NovoPar infere os dois tipos pelos argumentos, o que um literal composto não faz.
func NovoPar[A, B any](a A, b B) Par[A, B] { return Par[A, B]{a, b} }

func main() {
	fmt.Println("Soma(Ids):    ", Soma(Ids{1, 2, 3}), "| E inferido de S = Ids")
	fmt.Println("Soma([]float64):", Soma([]float64{0.5, 0.25}))

	pares := Filtrar([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
	fmt.Println("Filtrar:      ", pares)

	p := NovoPar("idade", 42)
	fmt.Printf("NovoPar:       %+v (%T)\n", p, p)
	q := Par[string, int]{"idade", 42} // no literal, os tipos precisam ser escritos
	fmt.Println("literal igual?", p == q)
}
