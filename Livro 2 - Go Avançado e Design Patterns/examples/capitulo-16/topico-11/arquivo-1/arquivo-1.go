// Command e13 reúne armadilhas de generics, executadas.
package main

import (
	"fmt"
	"io"
	"strings"
)

func Contem[T comparable](fatia []T, alvo T) bool {
	for _, v := range fatia {
		if v == alvo {
			return true
		}
	}
	return false
}

// Descrever é um falso genérico: o type switch sobre any(v) joga fora tudo que o T ofereceria.
func Descrever[T any](v T) string {
	switch x := any(v).(type) {
	case int:
		return fmt.Sprintf("inteiro %d", x)
	case string:
		return fmt.Sprintf("texto %q", x)
	default:
		return fmt.Sprintf("outro %v", x)
	}
}

// DescreverMelhor faz o mesmo sem generics.
func DescreverMelhor(v any) string {
	switch x := v.(type) {
	case int:
		return fmt.Sprintf("inteiro %d", x)
	case string:
		return fmt.Sprintf("texto %q", x)
	default:
		return fmt.Sprintf("outro %v", x)
	}
}

// LerTudoGenerica não ganha nada com o parâmetro de tipo: só chama métodos do io.Reader.
func LerTudoGenerica[R io.Reader](r R) (string, error) {
	b, err := io.ReadAll(r)
	return string(b), err
}

// LerTudo é a versão simples, que aceita qualquer io.Reader.
func LerTudo(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	return string(b), err
}

type Animal interface{ Som() string }

type Cao struct{}

func (Cao) Som() string { return "au" }

type Gato struct{}

func (Gato) Som() string { return "miau" }

// Uniao não existe em Go: Cao e Gato não são o "mesmo T".
func Primeiro[T any](a, b T) T { return a }

func main() {
	// 1. comparable aceita interfaces, e == entre dois valores não comparáveis dentro delas entra em pânico.
	func() {
		defer func() { fmt.Println("1. pânico em Contem[any]:", recover()) }()
		fmt.Println("1. Contem([]int,   1)    =", Contem([]int{1, 2}, 1))
		fmt.Println("   Contem([]any,   2)    =", Contem([]any{1, "a"}, any(2)))
		Contem([]any{[]int{1}}, any([]int{1})) // fatias dentro de any: == não é definido
	}()

	// 2. O falso genérico e o seu equivalente simples.
	fmt.Println("2.", Descrever(7), "|", Descrever("go"), "|", Descrever(2.5), "|", DescreverMelhor(7))

	// 3. Um parâmetro de tipo que só aceita um io.Reader não paga o seu custo de leitura.
	s1, _ := LerTudoGenerica(strings.NewReader("abc"))
	s2, _ := LerTudo(strings.NewReader("abc"))
	fmt.Println("3.", s1 == s2)

	// 4. T é UM tipo: Cao e Gato não se misturam em Primeiro, a menos que T seja a interface.
	fmt.Println("4.", Primeiro[Animal](Cao{}, Gato{}).Som())

	// 5. O zero de T: new(T) e var zero T dão o mesmo valor.
	var zero int
	fmt.Println("5.", *new(int) == zero, Contem([]int{0}, *new(int)))
}
