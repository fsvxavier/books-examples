package main

import (
	"fmt"

	"exemplos/livro2/capitulo-4/topico-5/projeto-1/banco"
)

func main() {
	p := banco.pool{id: 99}
	fmt.Println(p)
}
