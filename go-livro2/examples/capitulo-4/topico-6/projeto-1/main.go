package main

import (
	"fmt"
	"os"

	"exemplos/livro2/capitulo-4/topico-6/projeto-1/db"
)

var _ = db.Pool

func init() {
	fmt.Println("main: init()")
}

func main() {
	fmt.Println("main: comecou, args =", len(os.Args)-1)
}
