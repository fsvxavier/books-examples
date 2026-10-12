// Package calc tem funções simples para demonstrar testes.
package calc

import (
	"errors"
	"strings"
)

// ErrDivisaoPorZero indica uma divisão cujo divisor é zero.
var ErrDivisaoPorZero = errors.New("divisão por zero")

// Soma devolve a + b.
func Soma(a, b int) int { return a + b }

// Dividir devolve a / b, ou ErrDivisaoPorZero.
func Dividir(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivisaoPorZero
	}
	return a / b, nil
}

// Slug transforma um título em um trecho de URL.
func Slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), "-")
}
