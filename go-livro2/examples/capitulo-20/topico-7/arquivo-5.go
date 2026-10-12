package main

/*
#include <stdlib.h>
#include "lote.h"

extern int comparar(void *a, void *b);
extern void progresso(uintptr_t h, int percentual);

// qsort espera int (*)(const void *, const void *)
typedef int (*cmpfn)(const void *, const void *);
static void ordenar(int *v, int n) {
	qsort(v, n, sizeof(int), (cmpfn)comparar);
}
*/
import "C"

import (
	"fmt"
	"runtime/cgo"
	"unsafe"
)

//export comparar
func comparar(a, b unsafe.Pointer) C.int {
	x, y := *(*C.int)(a), *(*C.int)(b)
	switch {
	case x < y:
		return 1 // ordem decrescente de proposito
	case x > y:
		return -1
	}
	return 0
}

//export progresso
func progresso(h C.uintptr_t, percentual C.int) {
	f := cgo.Handle(h).Value().(func(int))
	f(int(percentual))
}

func main() {
	// 1. qsort do C chamando um comparador escrito em Go
	v := []C.int{5, 2, 9, 1, 7}
	C.ordenar(&v[0], C.int(len(v)))
	fmt.Println("ordenado pelo C com comparador Go:", v)

	// 2. callback de progresso: o closure vai por tras de um cgo.Handle
	var visto []int
	h := cgo.NewHandle(func(p int) { visto = append(visto, p) })
	defer h.Delete()
	n := C.processar(4, C.uintptr_t(h))
	fmt.Println("etapas:", n, "progresso visto pelo Go:", visto)
}
