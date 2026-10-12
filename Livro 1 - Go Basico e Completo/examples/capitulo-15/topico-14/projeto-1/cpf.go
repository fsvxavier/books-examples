// Package cpf valida e formata números de CPF.
package cpf

import (
	"errors"
	"fmt"

	"exemplo.com/cpf/internal/digitos"
)

// ErrInvalido indica um CPF que não passa na validação.
var ErrInvalido = errors.New("cpf inválido")

// Valido informa se s é um CPF válido, com ou sem pontuação.
func Valido(s string) bool {
	ds := digitos.Extrair(s)
	if len(ds) != 11 {
		return false
	}
	todosIguais := true
	for _, d := range ds[1:] {
		if d != ds[0] {
			todosIguais = false
			break
		}
	}
	if todosIguais {
		return false
	}
	return digitos.Verificador(ds[:9]) == ds[9] && digitos.Verificador(ds[:10]) == ds[10]
}

// Formatar devolve o CPF no formato 000.000.000-00.
func Formatar(s string) (string, error) {
	if !Valido(s) {
		return "", fmt.Errorf("%q: %w", s, ErrInvalido)
	}
	ds := digitos.Extrair(s)
	return fmt.Sprintf("%d%d%d.%d%d%d.%d%d%d-%d%d",
		ds[0], ds[1], ds[2], ds[3], ds[4], ds[5], ds[6], ds[7], ds[8], ds[9], ds[10]), nil
}
