// Command e02 mostra os tipos centrais (core types) e o padrão S ~[]E.
package main

import (
	"fmt"
	"slices"
)

// Ids é um tipo derivado de []int, com um método próprio.
type Ids []int

func (i Ids) Primeiro() int { return i[0] }

// ClonarSimples devolve sempre []E: o tipo derivado se perde.
func ClonarSimples[E any](fatia []E) []E {
	return append([]E(nil), fatia...)
}

// Clonar usa S ~[]E: S é o tipo da fatia (qualquer derivado de []E) e E é o elemento.
func Clonar[S ~[]E, E any](fatia S) S {
	return append(S(nil), fatia...)
}

// Chaves usa M ~map[K]V pela mesma razão: aceita mapas nomeados e devolve as chaves.
func Chaves[M ~map[K]V, K comparable, V any](m M) []K {
	saida := make([]K, 0, len(m))
	for k := range m {
		saida = append(saida, k)
	}
	return saida
}

// Tamanho funciona com uma união porque len é definido para todos os tipos dela.
func Tamanho[T ~string | ~[]byte](x T) int { return len(x) }

// Primeiro indexa: todos os tipos da união são indexáveis, com o mesmo tipo de resultado (byte).
func PrimeiroByte[T ~string | ~[]byte](x T) byte { return x[0] }

type Estoque map[string]int

func main() {
	ids := Ids{3, 1, 2}

	a := ClonarSimples(ids)
	b := Clonar(ids)
	fmt.Printf("ClonarSimples: %T\n", a)
	fmt.Printf("Clonar:        %T (o método Primeiro continua disponível: %d)\n", b, b.Primeiro())

	// A biblioteca padrão usa o mesmo padrão: slices.Clone, slices.Sort e outras preservam o tipo.
	c := slices.Clone(ids)
	slices.Sort(c)
	fmt.Printf("slices.Clone:  %T %v\n", c, c)

	fmt.Println("Chaves:", len(Chaves(Estoque{"a": 1, "b": 2})), "chaves")
	fmt.Println("Tamanho:", Tamanho("olá"), Tamanho([]byte{1, 2, 3}))
	fmt.Println("PrimeiroByte:", PrimeiroByte("A"), PrimeiroByte([]byte{66}))
}
