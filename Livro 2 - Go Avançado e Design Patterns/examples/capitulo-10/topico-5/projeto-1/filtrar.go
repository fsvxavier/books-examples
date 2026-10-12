package main

import "iter"

// Cada adaptador recebe uma Seq e devolve OUTRA Seq: nada e calculado ate alguem consumir.

func Filtrar[V any](seq iter.Seq[V], manter func(V) bool) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range seq {
			if manter(v) && !yield(v) {
				return
			}
		}
	}
}

func Mapear[V, R any](seq iter.Seq[V], f func(V) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Limitar entrega no maximo n elementos e PARA a origem logo depois do n-esimo.
func Limitar[V any](seq iter.Seq[V], n int) iter.Seq[V] {
	return func(yield func(V) bool) {
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

// Naturais e uma sequencia INFINITA: so existe porque a avaliacao e preguicosa.
func Naturais() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 1; ; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

// Zip junta duas sequencias em pares e termina quando a MENOR termina. Precisa de iter.Pull,
// porque e preciso avancar duas sequencias "no mesmo passo".
func Zip[A, B any](a iter.Seq[A], b iter.Seq[B]) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		proxA, paraA := iter.Pull(a)
		defer paraA()
		proxB, paraB := iter.Pull(b)
		defer paraB()
		for {
			x, ok1 := proxA()
			y, ok2 := proxB()
			if !ok1 || !ok2 || !yield(x, y) {
				return
			}
		}
	}
}

// Contar mostra quantos elementos a origem entregou (para provar a preguica).
func Contar[V any](seq iter.Seq[V], n *int) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range seq {
			*n++
			if !yield(v) {
				return
			}
		}
	}
}

// Reduzir consome tudo e devolve um valor (uma operacao "terminal", ao contrario das de cima).
func Reduzir[V, R any](seq iter.Seq[V], inicial R, f func(R, V) R) R {
	acc := inicial
	for v := range seq {
		acc = f(acc, v)
	}
	return acc
}
