// Command e01 explora os conjuntos de tipos (type sets) de uma constraint.
package main

import (
	"fmt"
	"strconv"
)

// Inteiro e Texto são conjuntos de tipos.
type Inteiro interface{ ~int | ~int32 | ~int64 }

type Texto interface{ ~string }

// Identificavel é uma interface só de métodos: o conjunto de TODOS os tipos que têm String.
type Identificavel interface{ String() string }

// Codigo é a interseção: tipos cujo tipo subjacente é int E que têm o método String.
// Embutir várias linhas em uma interface faz a INTERSEÇÃO dos conjuntos.
type Codigo interface {
	~int | ~int64
	String() string
}

// Etiqueta é um inteiro com String: pertence a Inteiro, a Identificavel e a Codigo.
type Etiqueta int

func (e Etiqueta) String() string { return "E" + strconv.Itoa(int(e)) }

// Pontos é um inteiro SEM String: pertence a Inteiro, mas não a Codigo.
type Pontos int

// Rotulo usa a interseção: soma 1 (operador de ~int) e mostra (método String).
func Rotulo[T Codigo](c T) string {
	return (c + 1).String()
}

func Dobrar[T Inteiro](x T) T { return x * 2 }

func main() {
	fmt.Println("Dobrar(Etiqueta(4)) =", Dobrar(Etiqueta(4)), "(tipo preservado, usa String)")
	fmt.Println("Dobrar(Pontos(4))   =", Dobrar(Pontos(4)))
	fmt.Println("Rotulo(Etiqueta(9)) =", Rotulo(Etiqueta(9)))
	var i Identificavel = Etiqueta(1)
	fmt.Println("Identificavel       =", i)
}
