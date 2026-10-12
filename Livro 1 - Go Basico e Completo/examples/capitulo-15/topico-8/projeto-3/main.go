package main

import (
	"fmt"

	"exemplo.com/dominio/pedido"
)

func main() {
	p := pedido.Pedido{ID: 1, Valor: 15000}
	fmt.Println(p.Resumo(), "| frete:", p.Frete())
}
