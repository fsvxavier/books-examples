package main

import "fmt"

//go:generate go tool stringer -type=Estado -linecomment
//go:generate go run ./gerador -saida=versao_gerada.go

// Estado é a etapa de um pedido.
type Estado int

const (
	Novo      Estado = iota // novo
	Pago                    // pago
	Enviado                 // enviado
	Entregue                // entregue
	Cancelado               // cancelado
)

func main() {
	fmt.Println(Pago, Entregue, Estado(9))
	fmt.Printf("%v %d %s\n", Enviado, Enviado, Versao)
}
