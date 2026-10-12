// Command e01 mostra reflect.Type, reflect.Value e a diferença entre tipo e espécie (Kind).
package main

import (
	"fmt"
	"reflect"
	"time"
)

// Reais é um tipo novo com espécie float64.
type Reais float64

// Usuario é uma struct comum.
type Usuario struct {
	Nome  string
	Idade int
}

// Saudador é uma interface.
type Saudador interface{ Saudar() string }

// Saudar faz de Usuario um Saudador.
func (u Usuario) Saudar() string { return "olá, " + u.Nome }

func descrever(x any) {
	t, v := reflect.TypeOf(x), reflect.ValueOf(x)
	if t == nil { // any(nil) não tem tipo
		fmt.Println("nil: TypeOf devolve nil, e ValueOf devolve um Value inválido:", !v.IsValid())
		return
	}
	literal, valor := fmt.Sprintf("%#v", x), fmt.Sprint(v)
	if k := t.Kind(); k == reflect.Func || k == reflect.Chan {
		literal, valor = "("+t.String()+")", "(endereço omitido)" // endereços mudam a cada execução
	}
	fmt.Printf("%-36s tipo=%-18s espécie=%-8s valor=%v\n", literal, t, t.Kind(), valor)
}

func main() {
	descrever(42)
	descrever(Reais(9.5))
	descrever("go")
	descrever([]int{1, 2})
	descrever([3]string{"a", "b", "c"})
	descrever(map[string]int{"a": 1})
	descrever(Usuario{"Ana", 30})
	descrever(&Usuario{"Bia", 25})
	descrever(time.Second)
	descrever(func(int) string { return "" })
	descrever(make(chan int))
	descrever(nil)

	fmt.Println()
	tUsuario := reflect.TypeFor[Usuario]()
	tSaudador := reflect.TypeFor[Saudador]()
	fmt.Println("Usuario implementa Saudador?", tUsuario.Implements(tSaudador))
	fmt.Println("*Usuario implementa Saudador?", reflect.PointerTo(tUsuario).Implements(tSaudador))
	fmt.Println("nome do tipo:", tUsuario.Name(), "| pacote:", tUsuario.PkgPath(), "| string:", tUsuario.String())
	fmt.Println("[]int: Name vazio?", reflect.TypeFor[[]int]().Name() == "", "| Elem:", reflect.TypeFor[[]int]().Elem())

	// Igualdade de tipos: Reais e float64 têm a mesma espécie, mas são tipos diferentes.
	fmt.Println("Reais == float64?", reflect.TypeFor[Reais]() == reflect.TypeFor[float64](), "| mesma espécie?",
		reflect.TypeFor[Reais]().Kind() == reflect.TypeFor[float64]().Kind())
}
