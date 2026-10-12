package main

import (
	"fmt"
	"iter"
)

// Linhas simula uma fonte com recurso a liberar (um arquivo, uma conexao): o defer DENTRO do
// iterador roda quando o laco termina, por qualquer motivo: fim, break ou panic do corpo.
func Linhas(nome string) iter.Seq[string] {
	return func(yield func(string) bool) {
		fmt.Println("  [abre", nome+"]")
		defer fmt.Println("  [fecha", nome+"]")
		for _, l := range []string{"a", "b", "c", "d"} {
			if !yield(l) {
				return
			}
		}
	}
}

func main() {
	fmt.Println("percorrendo tudo:")
	for l := range Linhas("f1") {
		fmt.Println("  linha", l)
	}

	fmt.Println("com break na linha b:")
	for l := range Linhas("f2") {
		fmt.Println("  linha", l)
		if l == "b" {
			break
		}
	}

	fmt.Println("com panic no corpo do laco:")
	func() {
		defer func() { fmt.Println("  recuperado:", recover()) }()
		for l := range Linhas("f3") {
			if l == "a" {
				panic("falha no corpo")
			}
		}
	}()

	fmt.Println("com return no meio de uma funcao:")
	func() {
		for l := range Linhas("f4") {
			if l == "c" {
				return
			}
		}
	}()
}
