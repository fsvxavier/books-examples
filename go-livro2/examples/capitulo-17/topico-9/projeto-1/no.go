// Package e07 tem uma árvore binária de busca genérica, com iteradores.
package e07

import (
	"cmp"
	"iter"
)

type no[K, V any] struct {
	chave    K
	valor    V
	esq, dir *no[K, V]
}

// Arvore é uma árvore binária de busca (sem balanceamento) de chaves K e valores V.
// A ordem vem de uma função de comparação, o que permite chaves de qualquer tipo.
type Arvore[K, V any] struct {
	raiz     *no[K, V]
	comparar func(a, b K) int
	tam      int
}

// Nova cria uma árvore para chaves ordenáveis (cmp.Ordered): usa cmp.Compare.
func Nova[K cmp.Ordered, V any]() *Arvore[K, V] {
	return &Arvore[K, V]{comparar: cmp.Compare[K]}
}

// NovaComFuncao cria uma árvore com a ordem dada, para chaves que não são cmp.Ordered.
func NovaComFuncao[K, V any](comparar func(a, b K) int) *Arvore[K, V] {
	return &Arvore[K, V]{comparar: comparar}
}

// Tamanho devolve o número de chaves.
func (a *Arvore[K, V]) Tamanho() int { return a.tam }

// Inserir guarda o valor; se a chave existe, troca o valor e devolve true.
func (a *Arvore[K, V]) Inserir(chave K, valor V) (substituiu bool) {
	pp := &a.raiz // ponteiro para o ponteiro: o mesmo código serve à raiz e aos filhos
	for *pp != nil {
		switch c := a.comparar(chave, (*pp).chave); {
		case c < 0:
			pp = &(*pp).esq
		case c > 0:
			pp = &(*pp).dir
		default:
			(*pp).valor = valor
			return true
		}
	}
	*pp = &no[K, V]{chave: chave, valor: valor}
	a.tam++
	return false
}

// Buscar devolve o valor da chave e se ela existe.
func (a *Arvore[K, V]) Buscar(chave K) (V, bool) {
	for n := a.raiz; n != nil; {
		switch c := a.comparar(chave, n.chave); {
		case c < 0:
			n = n.esq
		case c > 0:
			n = n.dir
		default:
			return n.valor, true
		}
	}
	var zero V
	return zero, false
}

// Todos percorre a árvore em ordem crescente de chave.
func (a *Arvore[K, V]) Todos() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		a.emOrdem(a.raiz, yield)
	}
}

// emOrdem devolve false quando quem consome pediu para parar.
func (a *Arvore[K, V]) emOrdem(n *no[K, V], yield func(K, V) bool) bool {
	if n == nil {
		return true
	}
	return a.emOrdem(n.esq, yield) && yield(n.chave, n.valor) && a.emOrdem(n.dir, yield)
}

// Intervalo percorre, em ordem, as chaves de min até max (inclusive), sem visitar
// os ramos que ficam fora do intervalo.
func (a *Arvore[K, V]) Intervalo(min, max K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		var visita func(n *no[K, V]) bool
		visita = func(n *no[K, V]) bool {
			if n == nil {
				return true
			}
			if a.comparar(min, n.chave) < 0 && !visita(n.esq) {
				return false
			}
			if a.comparar(n.chave, min) >= 0 && a.comparar(n.chave, max) <= 0 && !yield(n.chave, n.valor) {
				return false
			}
			if a.comparar(n.chave, max) < 0 {
				return visita(n.dir)
			}
			return true
		}
		visita(a.raiz)
	}
}

// Chaves devolve só as chaves, em ordem.
func (a *Arvore[K, V]) Chaves() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range a.Todos() {
			if !yield(k) {
				return
			}
		}
	}
}

// Altura é o comprimento do maior caminho da raiz a uma folha.
func (a *Arvore[K, V]) Altura() int {
	var h func(n *no[K, V]) int
	h = func(n *no[K, V]) int {
		if n == nil {
			return 0
		}
		return 1 + max(h(n.esq), h(n.dir))
	}
	return h(a.raiz)
}
