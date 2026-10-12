//go:build !cgo

package crc

import "hash/crc32"

// Implementacao informa qual versao esta compilada.
const Implementacao = "Go puro (hash/crc32)"

// Somar calcula o CRC-32 (IEEE) de dados, continuando a partir de inicial.
func Somar(inicial uint32, dados []byte) uint32 {
	return crc32.Update(inicial, crc32.IEEETable, dados)
}
