package main

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-1/capitulo-15/topico-8/projeto-3/pedido"
)

func main() {
	p := pedido.Pedido{ID: 1, Valor: 15000}
	fmt.Println(p.Resumo(), "| frete:", p.Frete())
}
