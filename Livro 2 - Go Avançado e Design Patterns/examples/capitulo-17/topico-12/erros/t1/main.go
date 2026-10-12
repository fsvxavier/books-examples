package main

import "fmt"

type Codigo interface {
	~int | ~int64
	String() string
}

// Pontos é um inteiro SEM String: está na união, mas falta o método.
type Pontos int

func Rotulo[T Codigo](c T) string { return (c + 1).String() }

func main() { fmt.Println(Rotulo(Pontos(1))) }
