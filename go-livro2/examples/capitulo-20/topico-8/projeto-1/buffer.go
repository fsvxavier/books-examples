package nativo

/*
#include <stdatomic.h>
#include <stdlib.h>

// o contador e tocado por goroutines diferentes (inclusive a do coletor de
// lixo), e o -race nao enxerga o C: por isso ele e atomico
static atomic_int vivos;

static void *reservar(size_t n) {
	void *p = calloc(1, n);
	if (p != NULL) atomic_fetch_add(&vivos, 1);
	return p;
}

static void devolver(void *p) {
	free(p);
	atomic_fetch_sub(&vivos, 1);
}

static int blocos_vivos(void) { return atomic_load(&vivos); }
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

// Buffer e memoria do heap do C, com dono unico: Close libera, e o coletor
// de lixo e a rede de seguranca caso alguem esqueca.
type Buffer struct {
	ptr     unsafe.Pointer
	n       int
	cleanup runtime.Cleanup
}

var ErrFechado = errors.New("buffer fechado")

func NovoBuffer(n int) (*Buffer, error) {
	p := C.reservar(C.size_t(n))
	if p == nil {
		return nil, errors.New("sem memoria")
	}
	b := &Buffer{ptr: p, n: n}
	b.cleanup = runtime.AddCleanup(b, func(p unsafe.Pointer) { C.devolver(p) }, p)
	return b, nil
}

// Bytes devolve uma VISAO da memoria do C, sem copiar. So vale ate Close.
func (b *Buffer) Bytes() []byte {
	if b.ptr == nil {
		return nil
	}
	return unsafe.Slice((*byte)(b.ptr), b.n)
}

// Copia devolve uma copia gerenciada pelo Go, que sobrevive ao Close.
func (b *Buffer) Copia() []byte {
	if b.ptr == nil {
		return nil
	}
	return C.GoBytes(b.ptr, C.int(b.n))
}

func (b *Buffer) Close() error {
	if b.ptr == nil {
		return ErrFechado
	}
	b.cleanup.Stop()
	C.devolver(b.ptr)
	b.ptr = nil
	return nil
}

func BlocosVivos() int { return int(C.blocos_vivos()) }
