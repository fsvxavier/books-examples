// Command e04 chama métodos e funções com reflexão.
package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Calculadora tem métodos exportados, um de ponteiro, um variádico e um não exportado.
type Calculadora struct{ historico []string }

// Somar soma dois inteiros.
func (c Calculadora) Somar(a, b int) int { return a + b }

// Dividir devolve um erro para divisor zero.
func (c Calculadora) Dividir(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("divisão por zero")
	}
	return a / b, nil
}

// Juntar é variádico.
func (c Calculadora) Juntar(sep string, partes ...string) string { return strings.Join(partes, sep) }

// Registrar tem receptor de ponteiro.
func (c *Calculadora) Registrar(msg string) { c.historico = append(c.historico, msg) }

func (c Calculadora) interno() int { return 1 }

func chamar(rotulo string, v reflect.Value, nome string, args ...any) {
	m := v.MethodByName(nome)
	if !m.IsValid() {
		fmt.Printf("%-22s método %q não existe neste Value\n", rotulo, nome)
		return
	}
	entradas := make([]reflect.Value, len(args))
	for i, a := range args {
		entradas[i] = reflect.ValueOf(a)
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%-22s PÂNICO: %v\n", rotulo, r)
		}
	}()
	saidas := m.Call(entradas)
	partes := make([]string, len(saidas))
	for i, s := range saidas {
		partes[i] = fmt.Sprint(s.Interface())
	}
	fmt.Printf("%-22s %s(%v) -> [%s]\n", rotulo, nome, args, strings.Join(partes, ", "))
}

func main() {
	var c Calculadora
	fmt.Println("Conjunto de métodos de Calculadora:")
	for m := range reflect.TypeFor[Calculadora]().Methods() {
		fmt.Printf("  %-10s %v\n", m.Name, m.Type)
	}
	fmt.Println("Conjunto de métodos de *Calculadora:", reflect.TypeFor[*Calculadora]().NumMethod(), "métodos")
	fmt.Println()

	vc := reflect.ValueOf(c)
	chamar("valor, Somar", vc, "Somar", 2, 3)
	chamar("valor, Dividir", vc, "Dividir", 7, 2)
	chamar("valor, Dividir por 0", vc, "Dividir", 7, 0)
	chamar("valor, Juntar", vc, "Juntar", "-", "a", "b", "c")
	chamar("valor, Registrar", vc, "Registrar", "x") // não está no conjunto de métodos de Calculadora
	chamar("valor, interno", vc, "interno")          // não exportado
	chamar("Somar, tipo errado", vc, "Somar", "2", 3)
	chamar("Somar, falta argumento", vc, "Somar", 2)

	vp := reflect.ValueOf(&c) // com o ponteiro, o método de ponteiro aparece
	chamar("ponteiro, Registrar", vp, "Registrar", "primeira")
	fmt.Println("historico:", c.historico)

	// Funções também: Call com um valor de função qualquer.
	f := reflect.ValueOf(strings.ToUpper)
	fmt.Println("\nCall em strings.ToUpper:", f.Call([]reflect.Value{reflect.ValueOf("go")})[0])
	fmt.Println("tipo:", f.Type(), "| entradas:", f.Type().NumIn(), "| saídas:", f.Type().NumOut())
}
