package main

/*
#include <errno.h>
#include <math.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#cgo LDFLAGS: -lm

typedef struct {
	int32_t id;
	double  preco;
	char    nome[16];
} Produto;

struct ponto {
	int x, y;
};

typedef union {
	int32_t  inteiro;
	float    real;
} Numero;

enum cor { VERMELHO, VERDE = 5, AZUL };

#define VERSAO "2.7"
#define LIMITE 100

static int dobro(int x) { return 2 * x; }
static void nada(void) {}
static int abrir(const char *caminho) {
	if (strlen(caminho) == 0) { errno = ENOENT; return -1; }
	return 3;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	// 1. tipos numericos: tamanho definido pelo C
	var i C.int = 7
	var l C.long = 7
	var u C.uint8_t = 200
	var d C.double = 2.5
	fmt.Printf("C.int=%d bytes  C.long=%d bytes  C.uint8_t=%d byte  C.double=%d bytes  int(Go)=%d bytes\n",
		unsafe.Sizeof(i), unsafe.Sizeof(l), unsafe.Sizeof(u), unsafe.Sizeof(d), unsafe.Sizeof(int(0)))
	fmt.Printf("tipos: %T %T %T %T\n", i, l, u, d)

	// 2. conversao explicita obrigatoria
	var n int = 10
	r := C.dobro(C.int(n))
	fmt.Println("dobro:", int(r))

	// 3. struct com typedef anonimo, struct nomeada, tamanho e offsets
	var p C.Produto
	p.id = 42
	p.preco = 19.9
	fmt.Println("Produto:", C.sizeof_Produto, "bytes; struct ponto:", C.sizeof_struct_ponto, "bytes")
	fmt.Println("campo nome e do tipo", fmt.Sprintf("%T", p.nome), "len", len(p.nome))
	pt := C.struct_ponto{x: 3, y: 4}
	fmt.Println("ponto:", pt.x, pt.y)

	// 4. union vira array de bytes
	var num C.Numero
	fmt.Printf("union: %T len=%d\n", num, len(num))
	*(*C.float)(unsafe.Pointer(&num)) = 1.0
	fmt.Printf("bits de 1.0f lidos como int32: %#x\n", *(*C.int32_t)(unsafe.Pointer(&num)))

	// 5. enum e #define
	fmt.Println("enum:", C.VERMELHO, C.VERDE, C.AZUL, "tipo", fmt.Sprintf("%T", C.AZUL))
	fmt.Println("define numerico:", C.LIMITE, "define texto:", C.VERSAO)

	// 6. errno como segundo retorno
	f, err := C.sqrt(-1)
	fmt.Println("sqrt(-1):", f, "err:", err)
	f2, err2 := C.sqrt(4)
	fmt.Println("sqrt(4):", f2, "err:", err2)

	vazio := C.CString("")
	defer C.free(unsafe.Pointer(vazio))
	fd, err3 := C.abrir(vazio)
	fmt.Println("abrir:", fd, "err:", err3)

	_, err4 := C.nada()
	fmt.Println("void:", err4)
}
