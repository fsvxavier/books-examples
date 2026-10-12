// Command e07 mostra por que um uintptr guardado em uma variável não acompanha o objeto: a pilha se move.
package main

import (
	"fmt"
	"unsafe"
)

//go:noinline
func crescerAPilha(n int) int {
	var folga [512]byte // um quadro grande por chamada: força a pilha a crescer
	folga[n%512] = byte(n)
	if n == 0 {
		return int(folga[0])
	}
	return crescerAPilha(n-1) + int(folga[n%512])
}

func main() {
	x := 1
	u := uintptr(unsafe.Pointer(&x)) // INVÁLIDO: guarda o endereço como inteiro
	crescerAPilha(2000)              // a pilha cresce e é copiada para outro lugar; x vai junto, u não
	x = 2
	p := (*int)(unsafe.Pointer(u)) // INVÁLIDO: converte de volta um endereço que já não é o de x
	fmt.Println("x =", x, "| lido pelo uintptr guardado =", *p)
	fmt.Println("o endereço guardado ainda é o de x?", u == uintptr(unsafe.Pointer(&x)))
}
