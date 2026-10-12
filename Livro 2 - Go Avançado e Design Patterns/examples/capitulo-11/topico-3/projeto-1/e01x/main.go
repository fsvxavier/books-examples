package main

import (
	"fmt"

	"exemplos/livro2/capitulo-11/topico-3/projeto-1/editor"
)

func main() {
	ed := editor.Novo("x")
	m := ed.Salvar()
	m.linhas = nil // o cliente tenta adulterar o instantâneo
	fmt.Println(m.titulo)
}
