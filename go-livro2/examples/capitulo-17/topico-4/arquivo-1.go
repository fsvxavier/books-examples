// Command e04 usa constraints recursivas: T se refere a si mesmo na própria constraint.
package main

import (
	"cmp"
	"fmt"
	"slices"
)

// Comparavel diz que T sabe se comparar com outro T. A constraint usa o próprio T.
type Comparavel[T any] interface {
	Comparar(outro T) int
}

// Versao é uma versão semântica simplificada.
type Versao struct{ Maior, Menor, Correcao int }

// Comparar ordena as versões.
func (v Versao) Comparar(o Versao) int {
	return cmp.Or(
		cmp.Compare(v.Maior, o.Maior),
		cmp.Compare(v.Menor, o.Menor),
		cmp.Compare(v.Correcao, o.Correcao),
	)
}

func (v Versao) String() string { return fmt.Sprintf("v%d.%d.%d", v.Maior, v.Menor, v.Correcao) }

// Maximo devolve o maior elemento. T Comparavel[T] é a constraint recursiva.
func Maximo[T Comparavel[T]](xs ...T) (T, bool) {
	var maior T
	if len(xs) == 0 {
		return maior, false
	}
	maior = xs[0]
	for _, x := range xs[1:] {
		if x.Comparar(maior) > 0 {
			maior = x
		}
	}
	return maior, true
}

// Clonavel diz que T sabe se copiar para outro T (um "construtor virtual").
type Clonavel[T any] interface{ Clonar() T }

// Perfil tem uma fatia: uma cópia rasa compartilharia o array.
type Perfil struct {
	Nome string
	Tags []string
}

// Clonar faz a cópia funda.
func (p Perfil) Clonar() Perfil {
	return Perfil{Nome: p.Nome, Tags: slices.Clone(p.Tags)}
}

// ClonarTodos copia uma lista de qualquer tipo clonável, e devolve a lista do MESMO tipo.
func ClonarTodos[T Clonavel[T]](xs []T) []T {
	saida := make([]T, len(xs))
	for i, x := range xs {
		saida[i] = x.Clonar()
	}
	return saida
}

func main() {
	vs := []Versao{{1, 2, 3}, {1, 10, 0}, {1, 2, 10}, {0, 9, 9}}
	m, _ := Maximo(vs...)
	fmt.Println("Maximo:", m)

	slices.SortFunc(vs, Versao.Comparar) // expressão de método: func(Versao, Versao) int
	fmt.Println("Ordenadas:", vs)

	orig := []Perfil{{"ana", []string{"a", "b"}}}
	copia := ClonarTodos(orig)
	copia[0].Tags[0] = "ALTERADA"
	fmt.Println("original:", orig[0].Tags, "| copia:", copia[0].Tags)
}
