package calc_test

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-1/capitulo-16/topico-2/projeto-1"
)

func ExampleSlug() {
	fmt.Println(calc.Slug("  Olá  Mundo Go "))
	// Output: olá-mundo-go
}

func ExampleDividir() {
	_, err := calc.Dividir(1, 0)
	fmt.Println(err)
	// Output: divisão por zero
}
