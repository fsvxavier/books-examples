package main

import (
	"fmt"

	_ "exemplos/livro2/capitulo-4/topico-12/projeto-1/cartao"
	"exemplos/livro2/capitulo-4/topico-12/projeto-1/pagamento"
)

func main() {
	fmt.Println("registrados:", pagamento.Nomes())
	_, err := pagamento.Novo("pix")
	fmt.Println("erro:", err)
}
