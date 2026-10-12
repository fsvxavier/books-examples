// Command e14 mostra que ordenar floats exige cuidado com NaN, mesmo com generics.
package main

import (
	"cmp"
	"fmt"
	"math"
	"slices"
)

func Maior[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func main() {
	nan := math.NaN()
	fmt.Println("Maior(NaN, 1)            =", Maior(nan, 1))
	fmt.Println("Maior(1, NaN)            =", Maior(1, nan))
	fmt.Println("cmp.Compare(NaN, 1)      =", cmp.Compare(nan, 1), "(NaN é menor que tudo)")
	fmt.Println("cmp.Compare(NaN, NaN)    =", cmp.Compare(nan, nan))
	fmt.Println("min(NaN, 1) / max(NaN, 1) =", min(nan, 1), max(nan, 1), "(os builtins propagam NaN)")

	xs := []float64{3, nan, 1, 2}
	slices.Sort(xs)
	fmt.Println("slices.Sort com NaN      =", xs, "(NaN vai para o começo)")
	fmt.Println("slices.Max com NaN       =", slices.Max([]float64{3, nan, 1}))

	defer func() { fmt.Println("slices.Max de fatia vazia: pânico:", recover()) }()
	_ = slices.Max([]int{})
}
