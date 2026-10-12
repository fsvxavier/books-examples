// Package zlibc e uma ligacao fina com a zlib do sistema.
package zlibc

/*
#cgo LDFLAGS: -lz
#cgo noescape compress2
#cgo nocallback compress2
#cgo noescape uncompress
#cgo nocallback uncompress
#include <zlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// Niveis aceitos por Comprimir.
const (
	Rapido = C.Z_BEST_SPEED
	Padrao = C.Z_DEFAULT_COMPRESSION
	Melhor = C.Z_BEST_COMPRESSION
)

var (
	ErrDadosCorrompidos = errors.New("zlibc: dados corrompidos")
	ErrTamanho          = errors.New("zlibc: tamanho original incorreto")
)

// Versao devolve a versao da biblioteca em uso.
func Versao() string { return C.GoString(C.zlibVersion()) }

// Comprimir devolve os dados no formato zlib (RFC 1950).
func Comprimir(dados []byte, nivel int) ([]byte, error) {
	limite := C.compressBound(C.uLong(len(dados)))
	destino := make([]byte, limite)
	n := limite

	var origem *C.Bytef
	if len(dados) > 0 {
		origem = (*C.Bytef)(unsafe.Pointer(&dados[0]))
	}
	rc := C.compress2((*C.Bytef)(unsafe.Pointer(&destino[0])), &n,
		origem, C.uLong(len(dados)), C.int(nivel))
	if rc != C.Z_OK {
		return nil, erro("compress2", rc)
	}
	return destino[:n], nil
}

// Descomprimir expande dados; tamanho e o tamanho original, que o chamador
// precisa guardar (a zlib nao o grava no formato).
func Descomprimir(dados []byte, tamanho int) ([]byte, error) {
	destino := make([]byte, max(tamanho, 1))
	n := C.uLong(tamanho)

	var origem *C.Bytef
	if len(dados) > 0 {
		origem = (*C.Bytef)(unsafe.Pointer(&dados[0]))
	}
	rc := C.uncompress((*C.Bytef)(unsafe.Pointer(&destino[0])), &n, origem, C.uLong(len(dados)))
	switch rc {
	case C.Z_OK:
		return destino[:n], nil
	case C.Z_BUF_ERROR:
		return nil, ErrTamanho
	case C.Z_DATA_ERROR:
		return nil, ErrDadosCorrompidos
	}
	return nil, erro("uncompress", rc)
}

func erro(funcao string, rc C.int) error {
	return fmt.Errorf("zlibc: %s falhou com codigo %d", funcao, int(rc))
}
