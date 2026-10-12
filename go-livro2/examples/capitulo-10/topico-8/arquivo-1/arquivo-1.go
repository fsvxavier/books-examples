package main

import (
	"fmt"
	"iter"
)

// iteradorRuim IGNORA o retorno de yield: continua produzindo depois do break.
func iteradorRuim() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 1; i <= 5; i++ {
			yield(i) // BUG: deveria ser `if !yield(i) { return }`
		}
	}
}

func main() {
	defer func() {
		fmt.Println("panic:", recover())
	}()
	for v := range iteradorRuim() {
		fmt.Println("recebi", v)
		if v == 2 {
			break
		}
	}
	fmt.Println("nao chega aqui")
}
