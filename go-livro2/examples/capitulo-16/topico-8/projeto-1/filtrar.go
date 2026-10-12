// Package e09 define combinadores genéricos sobre iteradores (iter.Seq).
package e09

import "iter"

// Filtrar produz só os valores para os quais manter é verdadeira.
func Filtrar[T any](seq iter.Seq[T], manter func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if manter(v) && !yield(v) {
				return
			}
		}
	}
}

// Mapear transforma cada valor.
func Mapear[T, R any](seq iter.Seq[T], f func(T) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Primeiros produz no máximo n valores e para: a fonte não é consumida além disso.
func Primeiros[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		i := 0
		for v := range seq {
			if !yield(v) {
				return
			}
			i++
			if i == n {
				return
			}
		}
	}
}

// Contar produz 0, 1, 2, ... sem fim.
func Contar() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; ; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

// Zip junta dois iteradores par a par, até o menor acabar.
func Zip[A, B any](a iter.Seq[A], b iter.Seq[B]) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		proxB, parar := iter.Pull(b)
		defer parar()
		for va := range a {
			vb, ok := proxB()
			if !ok || !yield(va, vb) {
				return
			}
		}
	}
}
