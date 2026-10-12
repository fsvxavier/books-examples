package main

import "fmt"

type Pedido struct {
	ID    int
	Valor int
}

func main() {
	bus := NovoBarramento()
	var log []string

	// Tres componentes que NAO se conhecem: estoque, financeiro e e-mail.
	bus.Inscrever("pedido.criado", func(e any) { log = append(log, fmt.Sprintf("estoque reserva itens do pedido %d", e.(Pedido).ID)) })
	cancelarFin := bus.Inscrever("pedido.criado", func(e any) { log = append(log, fmt.Sprintf("financeiro cobra %d", e.(Pedido).Valor)) })
	bus.Inscrever("pedido.criado", func(e any) { log = append(log, "e-mail avisa o cliente") })
	bus.Inscrever("pedido.cancelado", func(e any) { log = append(log, "estoque libera itens") })

	n, _ := bus.Publicar("pedido.criado", Pedido{1, 5000})
	fmt.Println("entregues:", n)
	for _, l := range log {
		fmt.Println("  ", l)
	}

	log = nil
	cancelarFin()
	cancelarFin() // idempotente
	n, _ = bus.Publicar("pedido.criado", Pedido{2, 700})
	fmt.Println("depois de cancelar o financeiro, entregues:", n, "| inscritos:", bus.Inscritos("pedido.criado"))

	n, _ = bus.Publicar("topico.sem.ninguem", nil)
	fmt.Println("topico sem inscritos, entregues:", n)

	// Um tratador que entra em panic nao derruba os outros.
	bus.Inscrever("x", func(any) { panic("bug no tratador") })
	ok := 0
	bus.Inscrever("x", func(any) { ok++ })
	n, err := bus.Publicar("x", nil)
	fmt.Println("com panic: entregues =", n, "| segundo tratador rodou:", ok == 1, "| erro:", err)

	// Um tratador pode se inscrever DENTRO de uma entrega, sem deadlock.
	bus.Inscrever("y", func(any) {
		bus.Inscrever("y", func(any) {})
	})
	n, _ = bus.Publicar("y", nil)
	fmt.Println("inscricao dentro de entrega: entregues =", n, "| inscritos agora:", bus.Inscritos("y"))
}
