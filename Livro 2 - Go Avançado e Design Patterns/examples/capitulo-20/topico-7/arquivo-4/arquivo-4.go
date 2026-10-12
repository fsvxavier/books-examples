package main

import "C"

import (
	"strings"
	"unicode/utf8"
)

//export SomarInteiros
func SomarInteiros(a, b C.int) C.int { return a + b }

// ContarRunas conta os caracteres (e nao os bytes) de uma string UTF-8 do C.
//
//export ContarRunas
func ContarRunas(s *C.char) C.int {
	return C.int(utf8.RuneCountInString(C.GoString(s)))
}

// Maiusculas devolve memoria alocada pelo C: quem chama precisa liberar com free.
//
//export Maiusculas
func Maiusculas(s *C.char) *C.char {
	return C.CString(strings.ToUpper(C.GoString(s)))
}

// main vazio e obrigatorio em -buildmode=c-shared.
func main() {}
