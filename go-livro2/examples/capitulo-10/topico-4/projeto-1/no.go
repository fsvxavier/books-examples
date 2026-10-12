package main

import "iter"

type No[T any] struct {
	Valor    T
	Esq, Dir *No[T]
}

type Arvore[T any] struct {
	Raiz    *No[T]
	Visitas int // conta quantos nos foram VISITADOS (para mostrar a avaliacao preguicosa)
}

// All percorre a arvore em profundidade (pre-ordem). Por convencao, o metodo se chama All.
// O iterador e uma FUNCAO que recebe o yield; o estado da travessia e a propria pilha de chamadas.
func (a *Arvore[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		a.profundidade(a.Raiz, yield)
	}
}

// profundidade devolve false quando o consumidor mandou parar: a parada sobe pela recursao.
func (a *Arvore[T]) profundidade(n *No[T], yield func(T) bool) bool {
	if n == nil {
		return true
	}
	a.Visitas++
	return yield(n.Valor) && a.profundidade(n.Esq, yield) && a.profundidade(n.Dir, yield)
}

// Largura percorre por niveis: "outro algoritmo de travessia" e so outro metodo.
func (a *Arvore[T]) Largura() iter.Seq[T] {
	return func(yield func(T) bool) {
		if a.Raiz == nil {
			return
		}
		fila := []*No[T]{a.Raiz}
		for len(fila) > 0 {
			n := fila[0]
			fila = fila[1:]
			a.Visitas++
			if !yield(n.Valor) {
				return
			}
			if n.Esq != nil {
				fila = append(fila, n.Esq)
			}
			if n.Dir != nil {
				fila = append(fila, n.Dir)
			}
		}
	}
}

// Folhas e Niveis mostram variacoes: filtrar durante a travessia, e entregar DOIS valores (Seq2).
func (a *Arvore[T]) Folhas() iter.Seq[T] {
	return func(yield func(T) bool) {
		var visitar func(n *No[T]) bool
		visitar = func(n *No[T]) bool {
			if n == nil {
				return true
			}
			if n.Esq == nil && n.Dir == nil {
				return yield(n.Valor)
			}
			return visitar(n.Esq) && visitar(n.Dir)
		}
		visitar(a.Raiz)
	}
}

func (a *Arvore[T]) ComNivel() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		var visitar func(n *No[T], nivel int) bool
		visitar = func(n *No[T], nivel int) bool {
			if n == nil {
				return true
			}
			return yield(nivel, n.Valor) && visitar(n.Esq, nivel+1) && visitar(n.Dir, nivel+1)
		}
		visitar(a.Raiz, 0)
	}
}
