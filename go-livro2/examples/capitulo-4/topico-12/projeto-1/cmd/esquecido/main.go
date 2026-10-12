package main

import (
	"fmt"

	_ "github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/cartao"
	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-12/projeto-1/pagamento"
)

func main() {
	fmt.Println("registrados:", pagamento.Nomes())
	_, err := pagamento.Novo("pix")
	fmt.Println("erro:", err)
}
