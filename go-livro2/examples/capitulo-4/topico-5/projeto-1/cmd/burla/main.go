package main

import (
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-5/projeto-1/banco"
)

func main() {
	p := banco.pool{id: 99}
	fmt.Println(p)
}
