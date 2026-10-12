package main

/*
#include <stdlib.h>

typedef struct {
	int *valores;
	int  n;
} Lote;

static int somar(Lote *l) {
	int s = 0;
	for (int i = 0; i < l->n; i++) s += l->valores[i];
	return s;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	vals := []C.int{1, 2, 3, 4}

	// o Lote mora no heap do C; o Go guarda nele um ponteiro para memoria do Go
	l := (*C.Lote)(C.malloc(C.size_t(C.sizeof_Lote)))
	defer C.free(unsafe.Pointer(l))
	l.valores = &vals[0]
	l.n = C.int(len(vals))

	fmt.Println("soma:", C.somar(l))
}
