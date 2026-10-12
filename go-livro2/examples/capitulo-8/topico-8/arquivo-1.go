package main

import (
	"fmt"
	"reflect"
	"time"
	"unique"
)

func main() {
	// 1) reflect.Type: um unico valor canonico por tipo (o ponteiro e comparavel).
	fmt.Println("reflect.TypeOf(1) == reflect.TypeOf(2):", reflect.TypeOf(1) == reflect.TypeOf(2))
	fmt.Println("reflect.TypeOf(1) == reflect.TypeOf(\"x\"):", reflect.TypeOf(1) == reflect.TypeOf("x"))

	// 2) *time.Location: instantes diferentes compartilham o mesmo ponteiro de fuso.
	t1, t2 := time.Now().UTC(), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println("mesma *Location em dois instantes:", t1.Location() == t2.Location())

	// 3) unique.Handle: igualdade de valores vira igualdade de ponteiros.
	a := unique.Make("sao paulo")
	b := unique.Make("sao " + "paulo")
	c := unique.Make("recife")
	fmt.Println("Handles de valores iguais:", a == b, "| de valores diferentes:", a == c)
}
