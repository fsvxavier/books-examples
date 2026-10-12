package main

import (
	"fmt"

	"exemplos/livro4/capitulo-2/topico-3/lib/internal/segredo"
	"exemplos/livro4/capitulo-2/topico-3/lib/palavra"
	"exemplos/livro4/capitulo-2/topico-3/lib/pkg/numero"
)

func main() {
	fmt.Println(palavra.Maiuscula("go"), numero.Dobro(21), segredo.Valor)
}
