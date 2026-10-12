package main

import (
	"fmt"
	"os"

	"github.com/exemplo/calc/internal/preco"
)

func main() {
	final, err := preco.ComDesconto(250.00, 15)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	fmt.Printf("Preço final: R$ %.2f\n", final)
}
