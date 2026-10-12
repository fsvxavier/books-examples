// Command e06 escreve em memória somente leitura através de uma string sem cópia.
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	s := "texto literal"                            // vive em uma região somente leitura do binário
	b := unsafe.Slice(unsafe.StringData(s), len(s)) // sem cópia: b aponta para essa região
	fmt.Println("lendo funciona:", string(b[:5]))
	fmt.Println("escrevendo em b[0]...")
	b[0] = 'T' // a escrita em memória somente leitura derruba o processo
	fmt.Println("não chega aqui")
}
