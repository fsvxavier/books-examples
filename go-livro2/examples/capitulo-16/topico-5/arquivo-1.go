// Command e05 apresenta tipos genéricos: pares, conjuntos, listas e aliases.
package main

import (
	"fmt"
	"sort"
	"strings"
)

// Par junta dois valores de tipos diferentes.
type Par[K comparable, V any] struct {
	Chave K
	Valor V
}

// String mostra o par como chave=valor.
func (p Par[K, V]) String() string { return fmt.Sprintf("%v=%v", p.Chave, p.Valor) }

// Conjunto é um conjunto de valores comparáveis.
type Conjunto[T comparable] map[T]struct{}

// NovoConjunto cria um conjunto com os valores dados.
func NovoConjunto[T comparable](valores ...T) Conjunto[T] {
	c := make(Conjunto[T], len(valores))
	for _, v := range valores {
		c.Adicionar(v)
	}
	return c
}

// Adicionar insere um valor.
func (c Conjunto[T]) Adicionar(v T) { c[v] = struct{}{} }

// Tem diz se o valor está no conjunto.
func (c Conjunto[T]) Tem(v T) bool {
	_, ok := c[v]
	return ok
}

// Uniao devolve um conjunto novo com os valores dos dois.
func (c Conjunto[T]) Uniao(outro Conjunto[T]) Conjunto[T] {
	r := make(Conjunto[T], len(c)+len(outro))
	for v := range c {
		r.Adicionar(v)
	}
	for v := range outro {
		r.Adicionar(v)
	}
	return r
}

// Lista é uma lista encadeada simples.
type Lista[T any] struct {
	cabeca *no[T]
	tam    int
}

type no[T any] struct {
	valor   T
	proximo *no[T]
}

// Inserir coloca um valor no início.
func (l *Lista[T]) Inserir(v T) {
	l.cabeca = &no[T]{valor: v, proximo: l.cabeca}
	l.tam++
}

// Valores devolve os valores em ordem, do início ao fim.
func (l *Lista[T]) Valores() []T {
	saida := make([]T, 0, l.tam)
	for n := l.cabeca; n != nil; n = n.proximo {
		saida = append(saida, n.valor)
	}
	return saida
}

// Dicionario é um alias genérico (Go 1.24): outro nome para um tipo parametrizado.
type Dicionario[V any] = map[string]V

// Contagem é um alias com um argumento já fixado.
type Contagem = Dicionario[int]

func main() {
	p := Par[string, int]{"idade", 42}
	fmt.Println(p, "-> tipo", fmt.Sprintf("%T", p))

	a := NovoConjunto(1, 2, 3)
	b := NovoConjunto(3, 4)
	u := a.Uniao(b)
	var itens []int
	for v := range u {
		itens = append(itens, v)
	}
	sort.Ints(itens)
	fmt.Println("união:", itens, "| tem 4?", u.Tem(4), "| tem 9?", u.Tem(9))

	var l Lista[string]
	for _, s := range strings.Fields("c b a") {
		l.Inserir(s)
	}
	fmt.Println("lista:", l.Valores())

	var c Contagem = Dicionario[int]{"go": 3}
	fmt.Printf("alias: %T, %v\n", c, c)

	// O zero de cada instanciação é o zero do seu tipo.
	var zp Par[string, *int]
	var zl Lista[float64]
	fmt.Printf("zeros: %v | %v | %d\n", zp, zl.Valores(), zl.tam)
}
