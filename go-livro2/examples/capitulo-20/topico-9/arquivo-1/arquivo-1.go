package nativo

/*
#include <stdlib.h>
#include <string.h>

#cgo nocallback somarC
#cgo nocallback somarVetorSemAnotacao
#cgo noescape somarVetorC
#cgo nocallback somarVetorC

static int somarC(int a, int b) { return a + b; }
static int vazia(void) { return 1; }

static long somarVetorC(int *v, int n) {
	long s = 0;
	for (int i = 0; i < n; i++) s += v[i];
	return s;
}
static long somarVetorSemAnotacao(int *v, int n) {
	long s = 0;
	for (int i = 0; i < n; i++) s += v[i];
	return s;
}

extern void ping(void);
static void chamar_ping(int n) {
	for (int i = 0; i < n; i++) ping();
}
*/
import "C"

import "unsafe"

// SomarGo e a versao em Go puro do que a chamada C faz.
//
//go:noinline
func SomarGo(a, b int) int { return a + b }

func SomarC(a, b int) int { return int(C.somarC(C.int(a), C.int(b))) }

func Vazia() int { return int(C.vazia()) }

func SomarVetorGo(v []int32) int64 {
	var s int64
	for _, x := range v {
		s += int64(x)
	}
	return s
}

func SomarVetorUmaChamada(v []int32) int64 {
	return int64(C.somarVetorC((*C.int)(unsafe.Pointer(&v[0])), C.int(len(v))))
}

func SomarVetorSemAnotacao(v []int32) int64 {
	return int64(C.somarVetorSemAnotacao((*C.int)(unsafe.Pointer(&v[0])), C.int(len(v))))
}

func SomarVetorUmaChamadaPorElemento(v []int32) int64 {
	var s int64
	for _, x := range v {
		s += int64(C.somarC(C.int(x), 0))
	}
	return s
}

func CStringIdaEVolta(s string) string {
	c := C.CString(s)
	defer C.free(unsafe.Pointer(c))
	return C.GoString(c)
}

var contagem int

//export ping
func ping() { contagem++ }

func ChamarPing(n int) { C.chamar_ping(C.int(n)) }
