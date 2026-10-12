package main

import (
	"fmt"

	nativo "exemplos/livro2/capitulo-20/topico-8/projeto-1"
)

func main() {
	b, _ := nativo.NovoBuffer(16)
	visao := b.Bytes()
	visao[0] = 7
	b.Close()
	fmt.Println("depois do Close, a visao ainda mostra:", visao[0])
}
