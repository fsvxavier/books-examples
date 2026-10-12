// Package pop conta bits ligados, para vermos o efeito de GOAMD64.
package pop

import "math/bits"

// Contar devolve quantos bits de x estão ligados.
//
//go:noinline
func Contar(x uint64) int { return bits.OnesCount64(x) }
