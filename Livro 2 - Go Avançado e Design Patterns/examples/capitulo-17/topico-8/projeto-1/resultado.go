// Package e06 define um tipo Resultado[T] e o compara com o estilo idiomático (T, error).
package e06

import (
	"errors"
	"fmt"
	"strconv"
)

// Resultado guarda um valor ou um erro, nunca os dois.
type Resultado[T any] struct {
	valor T
	err   error
}

// Ok cria um resultado com valor.
func Ok[T any](v T) Resultado[T] { return Resultado[T]{valor: v} }

// Falha cria um resultado com erro.
func Falha[T any](err error) Resultado[T] { return Resultado[T]{err: err} }

// De converte o par (T, error) de uma função comum em um Resultado.
func De[T any](v T, err error) Resultado[T] {
	if err != nil {
		return Falha[T](err)
	}
	return Ok(v)
}

// Desfazer devolve o par (T, error) de volta ao estilo idiomático.
func (r Resultado[T]) Desfazer() (T, error) { return r.valor, r.err }

// Padrao devolve o valor, ou o padrão se houve erro.
func (r Resultado[T]) Padrao(p T) T {
	if r.err != nil {
		return p
	}
	return r.valor
}

// Mapear transforma o valor, se houver, em um R. Um método genérico (Go 1.27).
func (r Resultado[T]) Mapear[R any](f func(T) R) Resultado[R] {
	if r.err != nil {
		return Falha[R](r.err)
	}
	return Ok(f(r.valor))
}

// Entao encadeia uma operação que também pode falhar.
func (r Resultado[T]) Entao[R any](f func(T) Resultado[R]) Resultado[R] {
	if r.err != nil {
		return Falha[R](r.err)
	}
	return f(r.valor)
}

// ErrNegativo indica um valor negativo.
var ErrNegativo = errors.New("valor negativo")

func raiz(n int) Resultado[int] {
	if n < 0 {
		return Falha[int](fmt.Errorf("raiz de %d: %w", n, ErrNegativo))
	}
	r := 0
	for (r+1)*(r+1) <= n {
		r++
	}
	return Ok(r)
}

// ComResultado converte um texto em inteiro, tira a raiz e formata, com Resultado.
func ComResultado(texto string) Resultado[string] {
	return De(strconv.Atoi(texto)).
		Entao(raiz).
		Mapear(func(n int) string { return "raiz=" + strconv.Itoa(n) })
}

// Idiomatico faz o mesmo no estilo comum de Go.
func Idiomatico(texto string) (string, error) {
	n, err := strconv.Atoi(texto)
	if err != nil {
		return "", err
	}
	if n < 0 {
		return "", fmt.Errorf("raiz de %d: %w", n, ErrNegativo)
	}
	r := 0
	for (r+1)*(r+1) <= n {
		r++
	}
	return "raiz=" + strconv.Itoa(r), nil
}
