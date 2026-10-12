// Command e15 testa os limites da inferência: o retorno não participa, e a instanciação parcial vale.
package main

import (
	"fmt"
	"strconv"
)

func Zero[T any]() T {
	var z T
	return z
}

func Mapear[T, R any](xs []T, f func(T) R) []R {
	var s []R
	for _, x := range xs {
		s = append(s, f(x))
	}
	return s
}

func main() {
	// O tipo de retorno esperado NÃO participa da inferência de uma chamada: é preciso dizer T.
	var x int = Zero[int]()
	fmt.Println("Zero[int]():", x)

	// Instanciação parcial: informa T e deixa R para o compilador.
	textos := Mapear[int](([]int{1, 2}), strconv.Itoa)
	fmt.Printf("Mapear[int](...): %q\n", textos)

	// Com uma função anônima, R vem do retorno do literal.
	dobros := Mapear([]float64{1.5, 2}, func(f float64) float64 { return f * 2 })
	fmt.Println("Mapear com literal:", dobros)
}
