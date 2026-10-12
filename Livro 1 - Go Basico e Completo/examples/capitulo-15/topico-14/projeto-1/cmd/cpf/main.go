// Comando cpf valida e formata os CPFs passados como argumentos.
package main

import (
	"fmt"
	"os"

	"exemplo.com/cpf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: cpf NUMERO...")
		os.Exit(2)
	}
	codigo := 0
	for _, arg := range os.Args[1:] {
		f, err := cpf.Formatar(arg)
		if err != nil {
			fmt.Println("invalido:", arg)
			codigo = 1
			continue
		}
		fmt.Println("valido:", f)
	}
	os.Exit(codigo)
}
