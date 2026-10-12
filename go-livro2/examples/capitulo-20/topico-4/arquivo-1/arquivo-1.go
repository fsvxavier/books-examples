package main

/*
#include <stdlib.h>
#include <string.h>

static size_t tamanho(const char *s) { return strlen(s); }

static void maiusculas(char *s) {
	for (; *s; s++) {
		if (*s >= 'a' && *s <= 'z') *s -= 32;
	}
}

// _GoString_: o C recebe a string do Go sem copia (so durante a chamada)
static size_t tamanho_go(_GoString_ s) { return _GoStringLen(s); }
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	original := "go e c"

	// CString copia: alterar a copia nao altera a string Go
	cs := C.CString(original)
	C.maiusculas(cs)
	fmt.Printf("copia em C: %q | original Go: %q\n", C.GoString(cs), original)

	// GoString tambem copia: depois do free, a string Go continua valida
	volta := C.GoString(cs)
	C.free(unsafe.Pointer(cs))
	fmt.Printf("depois do free, a string Go ainda vale: %q\n", volta)

	// byte nulo no meio: o C enxerga ate o NUL, o Go inteiro
	com := "ab\x00cd"
	c2 := C.CString(com)
	defer C.free(unsafe.Pointer(c2))
	fmt.Println("len em Go:", len(com), "| strlen em C:", C.tamanho(c2))
	fmt.Printf("GoString: %q | GoStringN(5): %q\n", C.GoString(c2), C.GoStringN(c2, 5))

	// _GoString_: sem copia, mas so o tamanho e o ponteiro durante a chamada
	fmt.Println("tamanho_go:", C.tamanho_go(com))

	// bytes: CBytes aloca no heap do C; GoBytes copia de volta
	dados := []byte{1, 2, 3, 4, 5}
	cb := C.CBytes(dados)
	defer C.free(cb)
	dados[0] = 99
	fmt.Println("original:", dados, "| copia em C:", C.GoBytes(cb, 5))
}
