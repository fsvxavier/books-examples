// Package montagem oferece um "builder de passos" generico: a construcao e uma lista de
// passos que modificam o produto, executados em ordem, com passos condicionais.
package montagem

import (
	"errors"
	"fmt"
)

// Passo modifica o produto em construcao e pode falhar.
type Passo[T any] func(*T) error

// Montar parte de um produto zerado, aplica os passos em ordem e devolve o resultado.
// Se um passo falha, a montagem para e o erro diz QUAL passo (pelo indice).
func Montar[T any](passos ...Passo[T]) (T, error) {
	var produto T
	for i, p := range passos {
		if p == nil {
			continue // passos nulos (ex.: Quando(false, ...)) sao ignorados
		}
		if err := p(&produto); err != nil {
			var zero T
			return zero, fmt.Errorf("passo %d: %w", i+1, err)
		}
	}
	return produto, nil
}

// Quando devolve o passo apenas se a condicao for verdadeira.
func Quando[T any](cond bool, p Passo[T]) Passo[T] {
	if cond {
		return p
	}
	return nil
}

// Validar e um passo final tipico: confere o produto ja montado.
func Validar[T any](f func(T) error) Passo[T] {
	return func(p *T) error { return f(*p) }
}

var ErrInvalido = errors.New("produto invalido")
