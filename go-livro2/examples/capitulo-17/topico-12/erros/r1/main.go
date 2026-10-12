package main

import "fmt"

type Comparavel[T any] interface{ Comparar(outro T) int }

type Item int

// O método compara com int, e não com Item.
func (i Item) Comparar(outro int) int { return int(i) - outro }

func Maximo[T Comparavel[T]](xs ...T) T {
	m := xs[0]
	for _, x := range xs[1:] {
		if x.Comparar(m) > 0 {
			m = x
		}
	}
	return m
}

func main() { fmt.Println(Maximo(Item(1), Item(2))) }
