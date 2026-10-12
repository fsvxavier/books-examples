//go:build cgo

package crc

/*
#cgo LDFLAGS: -lz
#cgo noescape crc32
#cgo nocallback crc32
#include <zlib.h>
*/
import "C"

import "unsafe"

// Implementacao informa qual versao esta compilada.
const Implementacao = "cgo (zlib)"

// Somar calcula o CRC-32 (IEEE) de dados, continuando a partir de inicial.
func Somar(inicial uint32, dados []byte) uint32 {
	if len(dados) == 0 {
		return inicial
	}
	r := C.crc32(C.ulong(inicial), (*C.Bytef)(unsafe.Pointer(&dados[0])), C.uInt(len(dados)))
	return uint32(r)
}
