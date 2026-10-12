// Command e07 olha os métodos genéricos pelo lado da reflexão.
package main

import (
	"fmt"
	"reflect"
)

type Lista[T any] struct{ itens []T }

func (l Lista[T]) Itens() []T { return l.itens }

func (l Lista[T]) Mapear[R any](f func(T) R) Lista[R] { return Lista[R]{} }

func main() {
	t := reflect.TypeFor[Lista[int]]()
	fmt.Println("métodos visíveis pela reflexão:", t.NumMethod())
	for i := range t.NumMethod() {
		fmt.Println(" -", t.Method(i).Name)
	}
}
