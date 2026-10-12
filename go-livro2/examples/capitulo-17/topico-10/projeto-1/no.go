// Package e08 tem uma árvore de prefixos (trie) genérica no valor.
package e08

import (
	"iter"
	"maps"
	"slices"
)

type no[V any] struct {
	filhos map[rune]*no[V]
	valor  V
	existe bool
}

// Trie associa textos a valores de tipo V e responde a buscas por prefixo.
type Trie[V any] struct {
	raiz no[V]
	tam  int
}

// Inserir associa o valor à chave; se a chave já existia, troca o valor.
func (t *Trie[V]) Inserir(chave string, valor V) {
	n := &t.raiz
	for _, r := range chave {
		if n.filhos == nil {
			n.filhos = make(map[rune]*no[V])
		}
		f, ok := n.filhos[r]
		if !ok {
			f = &no[V]{}
			n.filhos[r] = f
		}
		n = f
	}
	if !n.existe {
		t.tam++
	}
	n.valor, n.existe = valor, true
}

// Buscar devolve o valor da chave exata.
func (t *Trie[V]) Buscar(chave string) (V, bool) {
	n := t.achar(chave)
	if n == nil || !n.existe {
		var zero V
		return zero, false
	}
	return n.valor, true
}

func (t *Trie[V]) achar(prefixo string) *no[V] {
	n := &t.raiz
	for _, r := range prefixo {
		f, ok := n.filhos[r]
		if !ok {
			return nil
		}
		n = f
	}
	return n
}

// ComPrefixo percorre, em ordem alfabética, todas as chaves que começam por prefixo.
func (t *Trie[V]) ComPrefixo(prefixo string) iter.Seq2[string, V] {
	return func(yield func(string, V) bool) {
		n := t.achar(prefixo)
		if n == nil {
			return
		}
		percorrer(n, []rune(prefixo), yield)
	}
}

// percorrer devolve false quando quem consome pediu para parar.
func percorrer[V any](n *no[V], caminho []rune, yield func(string, V) bool) bool {
	if n.existe && !yield(string(caminho), n.valor) {
		return false
	}
	for _, r := range slices.Sorted(maps.Keys(n.filhos)) {
		if !percorrer(n.filhos[r], append(caminho, r), yield) {
			return false
		}
	}
	return true
}

// Tamanho devolve o número de chaves.
func (t *Trie[V]) Tamanho() int { return t.tam }
