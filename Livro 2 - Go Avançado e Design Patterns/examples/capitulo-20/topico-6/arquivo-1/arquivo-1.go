package main

/*
#include <stdlib.h>

typedef struct {
	int  *valores;
	int   n;
} Lote;

static int somar(Lote *l) {
	int s = 0;
	for (int i = 0; i < l->n; i++) s += l->valores[i];
	return s;
}

static int somar_vetor(int *v, int n) {
	int s = 0;
	for (int i = 0; i < n; i++) s += v[i];
	return s;
}
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"
)

func tentar(rotulo string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s: PANICO: %v\n", rotulo, r)
		}
	}()
	f()
}

func main() {
	vals := []C.int{1, 2, 3, 4}

	// 1. ponteiro Go para um vetor de inteiros: valido
	tentar("vetor", func() {
		fmt.Println("vetor:", C.somar_vetor(&vals[0], C.int(len(vals))))
	})

	// 2. ponteiro Go para struct que guarda OUTRO ponteiro Go: invalido
	tentar("struct", func() {
		l := C.Lote{valores: &vals[0], n: C.int(len(vals))}
		fmt.Println("struct:", C.somar(&l))
	})

	// 3. o mesmo, com o vetor preso (pinned) durante o uso
	tentar("pinner", func() {
		var p runtime.Pinner
		p.Pin(&vals[0])
		defer p.Unpin()
		l := C.Lote{valores: &vals[0], n: C.int(len(vals))}
		fmt.Println("pinner:", C.somar(&l))
	})

	// 4. alternativa sem Go pointer dentro do Go pointer: o vetor vem do C
	tentar("malloc", func() {
		m := (*[4]C.int)(C.malloc(C.size_t(4 * C.sizeof_int)))
		defer C.free(unsafe.Pointer(m))
		for i := range m {
			m[i] = C.int(i + 1)
		}
		l := C.Lote{valores: &m[0], n: 4}
		fmt.Println("malloc:", C.somar(&l))
	})

	if len(os.Args) > 1 {
		fmt.Println("(modo", os.Args[1], ")")
	}
}
