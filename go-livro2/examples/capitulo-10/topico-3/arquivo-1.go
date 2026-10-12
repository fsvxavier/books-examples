package main

import "fmt"

// ---------- A forma CLASSICA do catalogo: uma interface com HasNext/Next ----------

type Iterador[T any] interface {
	TemProximo() bool
	Proximo() T
}

// Colecao cria iteradores compativeis com ela.
type Colecao[T any] interface {
	CriarIterador() Iterador[T]
}

// ---------- Colecao concreta: uma arvore binaria ----------

type No[T any] struct {
	Valor    T
	Esq, Dir *No[T]
}

type Arvore[T any] struct{ Raiz *No[T] }

// Iterador em profundidade (pre-ordem): guarda o ESTADO da travessia (uma pilha) no proprio objeto.
type iteradorProfundidade[T any] struct{ pilha []*No[T] }

func (a *Arvore[T]) CriarIterador() Iterador[T] {
	it := &iteradorProfundidade[T]{}
	if a.Raiz != nil {
		it.pilha = append(it.pilha, a.Raiz)
	}
	return it
}

func (it *iteradorProfundidade[T]) TemProximo() bool { return len(it.pilha) > 0 }

func (it *iteradorProfundidade[T]) Proximo() T {
	n := it.pilha[len(it.pilha)-1]
	it.pilha = it.pilha[:len(it.pilha)-1]
	if n.Dir != nil { // empilha a direita primeiro, para a esquerda sair antes
		it.pilha = append(it.pilha, n.Dir)
	}
	if n.Esq != nil {
		it.pilha = append(it.pilha, n.Esq)
	}
	return n.Valor
}

// Iterador em largura: OUTRO algoritmo de travessia da MESMA colecao (uma fila).
type iteradorLargura[T any] struct{ fila []*No[T] }

func (a *Arvore[T]) IteradorLargura() Iterador[T] {
	it := &iteradorLargura[T]{}
	if a.Raiz != nil {
		it.fila = append(it.fila, a.Raiz)
	}
	return it
}

func (it *iteradorLargura[T]) TemProximo() bool { return len(it.fila) > 0 }

func (it *iteradorLargura[T]) Proximo() T {
	n := it.fila[0]
	it.fila = it.fila[1:]
	if n.Esq != nil {
		it.fila = append(it.fila, n.Esq)
	}
	if n.Dir != nil {
		it.fila = append(it.fila, n.Dir)
	}
	return n.Valor
}

// O cliente trabalha SO com a interface: serve para qualquer colecao e qualquer algoritmo.
func Imprimir[T any](rotulo string, it Iterador[T]) {
	fmt.Print(rotulo, ":")
	for it.TemProximo() {
		fmt.Print(" ", it.Proximo())
	}
	fmt.Println()
}

func main() {
	//        1
	//      /   \
	//     2     3
	//    / \     \
	//   4   5     6
	a := &Arvore[int]{&No[int]{1,
		&No[int]{2, &No[int]{Valor: 4}, &No[int]{Valor: 5}},
		&No[int]{3, nil, &No[int]{Valor: 6}},
	}}

	Imprimir("profundidade", a.CriarIterador())
	Imprimir("largura     ", a.IteradorLargura())

	// Dois iteradores sobre a MESMA colecao, ao mesmo tempo, cada um com o seu estado.
	x, y := a.CriarIterador(), a.IteradorLargura()
	fmt.Print("intercalados:")
	for x.TemProximo() && y.TemProximo() {
		fmt.Print(" ", x.Proximo(), "/", y.Proximo())
	}
	fmt.Println()
}
