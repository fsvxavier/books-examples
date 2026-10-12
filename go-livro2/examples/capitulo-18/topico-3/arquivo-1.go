// Command e03 mostra quando um reflect.Value pode ser alterado, e os pânicos de quando não pode.
package main

import (
	"fmt"
	"reflect"
)

// Conta tem um campo exportado e outro não.
type Conta struct {
	Saldo int
	nome  string
}

// tentar roda f e mostra o pânico, se houver.
func tentar(rotulo string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%-34s PÂNICO: %v\n", rotulo, r)
		}
	}()
	f()
}

func main() {
	c := Conta{Saldo: 10, nome: "ana"}

	// 1. ValueOf(c) recebe uma CÓPIA: o Value não é endereçável.
	v := reflect.ValueOf(c)
	fmt.Println("ValueOf(c):   CanAddr =", v.CanAddr(), "| Field(0).CanSet =", v.Field(0).CanSet())
	tentar("SetInt em uma cópia", func() { v.Field(0).SetInt(99) })

	// 2. ValueOf(&c).Elem() é a própria variável: endereçável e alterável.
	e := reflect.ValueOf(&c).Elem()
	fmt.Println("ValueOf(&c).Elem(): CanAddr =", e.CanAddr(), "| Saldo.CanSet =", e.Field(0).CanSet(), "| nome.CanSet =", e.Field(1).CanSet())
	e.Field(0).SetInt(99)
	fmt.Println("depois de SetInt(99): c.Saldo =", c.Saldo)

	// 3. Campo não exportado: visível, mas não alterável.
	tentar("Set em campo não exportado", func() { e.Field(1).SetString("bia") })

	// 4. Tipos errados: SetInt em uma string, Set com tipo incompatível.
	tentar("SetInt em uma string", func() { reflect.ValueOf(new(string)).Elem().SetInt(1) })
	tentar("Set com tipo incompatível", func() { e.Field(0).Set(reflect.ValueOf("texto")) })

	// 5. Elem em algo que não é ponteiro nem interface.
	tentar("Elem em um int", func() { reflect.ValueOf(42).Elem() })

	// 6. Ponteiro nulo: Elem devolve um Value zero, e usá-lo entra em pânico.
	var p *Conta
	ev := reflect.ValueOf(p).Elem()
	fmt.Println("Elem de *Conta nil: IsValid =", ev.IsValid())
	tentar("Field em Value inválido", func() { ev.Field(0) })

	// 7. Overflow: SetInt não trunca nem avisa; use OverflowInt antes.
	var pequeno int8
	pv := reflect.ValueOf(&pequeno).Elem()
	fmt.Println("int8: OverflowInt(300) =", pv.OverflowInt(300))
	pv.SetInt(300)
	fmt.Println("depois de SetInt(300): pequeno =", pequeno, "(truncado em silêncio)")

	// 8. Mapas e fatias: SetMapIndex e Append funcionam sem endereço, porque o valor já é uma referência.
	m := map[string]int{}
	reflect.ValueOf(m).SetMapIndex(reflect.ValueOf("x"), reflect.ValueOf(5))
	fmt.Println("SetMapIndex:", m)
	s := []int{1, 2, 3}
	reflect.ValueOf(s).Index(0).SetInt(100) // os elementos de uma fatia são endereçáveis
	fmt.Println("Index(0).SetInt:", s)
}
