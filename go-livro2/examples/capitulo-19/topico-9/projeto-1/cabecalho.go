// Package e09 fixa o tamanho e as posições de uma struct em tempo de compilação.
package e09

import "unsafe"

// Cabecalho é escrito e lido como bytes por outro sistema: o layout é um contrato.
type Cabecalho struct {
	Magico  uint32
	Versao  uint16
	Flags   uint16
	Tamanho uint64
}

// As verificações abaixo não geram código: se o contrato for violado, o programa NÃO COMPILA.
// Um array de tamanho negativo é um erro de compilação; "x - y" como tamanho só é válido se x >= y.
const (
	tamanhoEsperado = 16
	_               = uint(unsafe.Sizeof(Cabecalho{}) - tamanhoEsperado) // falha se Sizeof < 16
	_               = uint(tamanhoEsperado - unsafe.Sizeof(Cabecalho{})) // falha se Sizeof > 16
	_               = uint(unsafe.Offsetof(Cabecalho{}.Tamanho) - 8)
	_               = uint(8 - unsafe.Offsetof(Cabecalho{}.Tamanho))
)
