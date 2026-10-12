package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

// Contagem devolve um iterador que produz os números de 1 até n.
func Contagem(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 1; i <= n; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func main() {
	nomes := []string{"Ana", "Bruno", "Carla"}
	for i, nome := range nomes {
		fmt.Printf("%d:%s ", i, nome)
	}
	fmt.Println()

	estoque := map[string]int{"maçã": 3, "banana": 7, "uva": 1}
	for _, k := range slices.Sorted(maps.Keys(estoque)) {
		fmt.Printf("%s=%d ", k, estoque[k])
	}
	fmt.Println()

	for i, r := range "café" {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()

	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	close(ch)
	for v := range ch {
		fmt.Print(v, " ")
	}
	fmt.Println()

	for v := range Contagem(5) {
		if v == 4 {
			break
		}
		fmt.Print(v, " ")
	}
	fmt.Println()

	for i, v := range slices.Backward(nomes) {
		fmt.Printf("%d:%s ", i, v)
	}
	fmt.Println()
}
