package main

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-5/projeto-1/banco"
)

func main() {
	a, b := banco.Obter(), banco.Obter()
	fmt.Println("mesma instancia:", a == b, "id:", a.ID())
	fmt.Println(a.Consultar("SELECT 1"))

	// Com tipo exportado, o "singleton" vira sugestao: nada impede um segundo.
	x, y := banco.PoolAberto{ID: 1}, banco.PoolAberto{ID: 2}
	fmt.Println("PoolAberto iguais?", x == y)
}
