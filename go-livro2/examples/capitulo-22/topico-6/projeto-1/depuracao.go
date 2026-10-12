// Package cte mostra constantes, código morto e divisão por constante.
package cte

const depuracao = false

// Registra só faz algo se a constante for verdadeira.
//
//go:noinline
func Registra(n int) int {
	if depuracao {
		println("n =", n)
	}
	return n * 2
}

// DivideDez divide por uma constante.
//
//go:noinline
func DivideDez(n uint32) uint32 { return n / 10 }

// Divide divide por um valor desconhecido na compilação.
//
//go:noinline
func Divide(n, d uint32) uint32 { return n / d }

// Multiplo8 usa módulo de potência de dois.
//
//go:noinline
func Multiplo8(n uint32) uint32 { return n % 8 }
