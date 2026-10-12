// Command e04 mostra a inferência de tipos, inclusive a generalização do Go 1.27.
package main

import (
	"fmt"
	"slices"
	"strconv"
)

// Mapear aplica f a cada elemento.
func Mapear[T, R any](fatia []T, f func(T) R) []R {
	saida := make([]R, 0, len(fatia))
	for _, v := range fatia {
		saida = append(saida, f(v))
	}
	return saida
}

// Somar soma dois valores do mesmo tipo.
func Somar[T ~int | ~float64](a, b T) T { return a + b }

// Dobro multiplica por dois.
func Dobro[T ~int | ~float64](x T) T { return x * 2 }

type Aplicador func(int) int

func main() {
	// 1. Pelos argumentos.
	fmt.Println(Somar(1, 2), Somar(1.5, 2.5))

	// 2. Constantes sem tipo: com tipos misturados, vale o "maior" tipo padrão (Go 1.21).
	fmt.Printf("Somar(1, 2.5) = %v (%T)\n", Somar(1, 2.5), Somar(1, 2.5))

	// 3. Por uma função passada como argumento: R é inferido do retorno de strconv.Itoa.
	textos := Mapear([]int{1, 2, 3}, strconv.Itoa)
	fmt.Printf("Mapear(..., strconv.Itoa) = %q\n", textos)

	// 4. Uma função genérica sem instanciar, passada onde se espera uma função comum.
	dobrados := Mapear([]int{1, 2, 3}, Dobro)
	fmt.Println("Mapear(..., Dobro)       =", dobrados)

	// 5. Go 1.27: a inferência vale ao ATRIBUIR a função genérica a uma variável de tipo função.
	var f func(float64) float64 = Dobro
	var g Aplicador = Dobro
	fmt.Println("var f func(float64) float64 = Dobro  ->", f(1.25))
	fmt.Println("var g Aplicador = Dobro              ->", g(21))

	// 6. Também na conversão e ao passar para slices.SortFunc.
	h := Aplicador(Dobro)
	fmt.Println("Aplicador(Dobro)                     ->", h(4))
	nomes := []string{"maria", "ana", "joao"}
	slices.SortFunc(nomes, func(a, b string) int { return len(a) - len(b) })
	fmt.Println("slices.SortFunc                      ->", nomes)
}
