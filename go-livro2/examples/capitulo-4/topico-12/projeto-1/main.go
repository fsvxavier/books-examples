package main

import (
	"fmt"

	_ "github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/boleto"
	_ "github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/cartao"
	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/pagamento"
	_ "github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/pix"
)

func main() {
	fmt.Println("registrados:", pagamento.Nomes())
	m, err := pagamento.Novo("pix")
	if err != nil {
		fmt.Println("erro:", err)
		return
	}
	msg, _ := m.Cobrar(15000)
	fmt.Println(msg)
}
