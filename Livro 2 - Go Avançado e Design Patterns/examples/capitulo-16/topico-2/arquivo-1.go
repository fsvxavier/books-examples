// Command e02 apresenta funções genéricas e constraints.
package main

import (
	"cmp"
	"fmt"
	"strings"
)

// Maior devolve o maior de dois valores de qualquer tipo ordenável.
func Maior[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Numero é uma constraint: qualquer tipo cujo tipo subjacente seja um destes.
type Numero interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64
}

// Soma soma os valores de uma fatia.
func Soma[T Numero](valores []T) T {
	var total T
	for _, v := range valores {
		total += v
	}
	return total
}

// Contem diz se a fatia tem o valor. comparable permite == e !=.
func Contem[T comparable](fatia []T, alvo T) bool {
	for _, v := range fatia {
		if v == alvo {
			return true
		}
	}
	return false
}

// Primeiro devolve o primeiro elemento, ou o zero de T e false se a fatia é vazia.
func Primeiro[T any](fatia []T) (T, bool) {
	if len(fatia) == 0 {
		var zero T
		return zero, false
	}
	return fatia[0], true
}

// Reais é um tipo novo, cujo tipo subjacente é float64.
type Reais float64

func (r Reais) String() string { return fmt.Sprintf("R$ %.2f", float64(r)) }

func main() {
	fmt.Println("Maior(3, 7)          =", Maior(3, 7))
	fmt.Println("Maior(2.5, 1.5)      =", Maior(2.5, 1.5))
	fmt.Println(`Maior("go", "rust")  =`, Maior("go", "rust"))
	fmt.Println("Maior[float64](3, 7) =", Maior[float64](3, 7), "(argumento de tipo explícito)")

	fmt.Println("Soma([]int)          =", Soma([]int{1, 2, 3}))
	fmt.Println("Soma([]float64)      =", Soma([]float64{0.5, 0.25}))
	precos := []Reais{10.5, 20.25}
	fmt.Println("Soma([]Reais)        =", Soma(precos), "(o ~ aceita tipos novos)")

	fmt.Println("Contem               =", Contem([]string{"a", "b"}, "b"), Contem([]int{1, 2}, 9))
	if _, ok := Primeiro([]string(nil)); !ok {
		fmt.Println("Primeiro de vazia    = zero e false")
	}
	v, _ := Primeiro([]string{"x", "y"})
	fmt.Printf("Primeiro de ([x y])  = %q\n", v)

	// O tipo é fixado na compilação: cada instanciação é um tipo de função diferente.
	fmt.Printf("tipos: %T e %T\n", Maior[int], Maior[string])
	fmt.Println(strings.Repeat("-", 20))
}
