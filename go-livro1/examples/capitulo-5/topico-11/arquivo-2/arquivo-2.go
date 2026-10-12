package main

import (
	"cmp"
	"fmt"
)

// No é um nó da árvore binária de busca.
type No struct {
	Valor    int
	Esq, Dir *No
}

// Arvore guarda só a raiz. O valor zero é uma árvore vazia e já pode ser usada.
type Arvore struct {
	raiz    *No
	tamanho int
}

func (a *Arvore) Inserir(v int) bool {
	var inserido bool
	a.raiz, inserido = inserir(a.raiz, v)
	if inserido {
		a.tamanho++
	}
	return inserido
}

// inserir devolve a raiz (possivelmente nova) da subárvore.
func inserir(n *No, v int) (*No, bool) {
	if n == nil {
		return &No{Valor: v}, true
	}
	var inserido bool
	switch cmp.Compare(v, n.Valor) {
	case -1:
		n.Esq, inserido = inserir(n.Esq, v)
	case 1:
		n.Dir, inserido = inserir(n.Dir, v)
	}
	return n, inserido
}

func (a *Arvore) Contem(v int) bool {
	n := a.raiz
	for n != nil {
		switch c := cmp.Compare(v, n.Valor); {
		case c < 0:
			n = n.Esq
		case c > 0:
			n = n.Dir
		default:
			return true
		}
	}
	return false
}

// EmOrdem devolve os valores em ordem crescente.
func (a *Arvore) EmOrdem() []int {
	var saida []int
	var visitar func(*No)
	visitar = func(n *No) {
		if n == nil {
			return
		}
		visitar(n.Esq)
		saida = append(saida, n.Valor)
		visitar(n.Dir)
	}
	visitar(a.raiz)
	return saida
}

func (a *Arvore) Altura() int { return altura(a.raiz) }

func altura(n *No) int {
	if n == nil {
		return 0
	}
	return 1 + max(altura(n.Esq), altura(n.Dir))
}

func main() {
	var a Arvore
	for _, v := range []int{50, 30, 70, 20, 40, 60, 80, 30} {
		if !a.Inserir(v) {
			fmt.Println("repetido, ignorado:", v)
		}
	}
	fmt.Println(a.EmOrdem(), a.tamanho, a.Altura())
	fmt.Println(a.Contem(60), a.Contem(65))

	var vazia Arvore
	fmt.Println(vazia.EmOrdem(), vazia.Altura(), vazia.Contem(1))
}
