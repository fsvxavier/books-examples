package main

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-3/lib/internal/segredo"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-3/lib/palavra"
	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-2/topico-3/lib/pkg/numero"
)

func main() {
	fmt.Println(palavra.Maiuscula("go"), numero.Dobro(21), segredo.Valor)
}
