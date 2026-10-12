package main

/*
static void guardar(void *p) {}
*/
import "C"

import "unsafe"

type Lote struct {
	nome  *string
	itens []int
}

func main() {
	nome := "x"
	l := Lote{nome: &nome}
	C.guardar(unsafe.Pointer(&l))
	C.guardar(unsafe.Pointer(&l.itens))
	var n int
	C.guardar(unsafe.Pointer(&n))
}
