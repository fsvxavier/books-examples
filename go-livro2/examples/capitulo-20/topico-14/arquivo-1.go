// Package varredura mostra um trabalho longo em C com progresso e
// cancelamento.
package varredura

/*
#include "varredura.h"

// declaracao da funcao Go exportada; a definicao fica em varredura.c
extern int avancar(uintptr_t h, int percentual);
*/
import "C"

import (
	"context"
	"errors"
	"runtime/cgo"
	"unsafe"
)

var ErrCancelado = errors.New("varredura cancelada")

type estado struct {
	ctx       context.Context
	aoAvancar func(percentual int)
}

//export avancar
func avancar(h C.uintptr_t, percentual C.int) C.int {
	e := cgo.Handle(h).Value().(*estado)
	if e.aoAvancar != nil {
		e.aoAvancar(int(percentual))
	}
	if e.ctx.Err() != nil {
		return 1
	}
	return 0
}

// Somar soma os bytes de dados no C. O contexto e consultado a cada etapa;
// aoAvancar (opcional) recebe 10, 20, ..., 100.
func Somar(ctx context.Context, dados []byte, aoAvancar func(int)) (int64, error) {
	if len(dados) == 0 {
		return 0, nil
	}
	h := cgo.NewHandle(&estado{ctx: ctx, aoAvancar: aoAvancar})
	defer h.Delete()

	r := C.varrer((*C.uchar)(unsafe.Pointer(&dados[0])), C.size_t(len(dados)), C.uintptr_t(h))
	if r < 0 {
		return 0, context.Cause(ctx)
	}
	return int64(r), nil
}
