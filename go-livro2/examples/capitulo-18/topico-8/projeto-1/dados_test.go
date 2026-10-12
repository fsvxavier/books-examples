package e08

import (
	"reflect"
	"testing"
)

type Dados struct {
	A, B int
	Nome string
}

var sinkInt int
var sinkStr string

func BenchmarkDireto(b *testing.B) {
	d := &Dados{A: 1, B: 42}
	for b.Loop() {
		sinkInt = d.B
	}
}

func BenchmarkReflexaoPorNome(b *testing.B) {
	d := &Dados{A: 1, B: 42}
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = int(reflect.ValueOf(d).Elem().FieldByName("B").Int())
	}
}

func BenchmarkReflexaoPorIndice(b *testing.B) {
	d := &Dados{A: 1, B: 42}
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = int(reflect.ValueOf(d).Elem().Field(1).Int())
	}
}

// Com o Value guardado (a situação de um validador ou de um decodificador que prepara tudo antes).
func BenchmarkReflexaoValueGuardado(b *testing.B) {
	d := &Dados{A: 1, B: 42}
	campo := reflect.ValueOf(d).Elem().Field(1)
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = int(campo.Int())
	}
}

func BenchmarkEscritaDireta(b *testing.B) {
	d := &Dados{}
	for b.Loop() {
		d.B = 7
	}
}

func BenchmarkEscritaReflexao(b *testing.B) {
	d := &Dados{}
	campo := reflect.ValueOf(d).Elem().Field(1)
	b.ReportAllocs()
	for b.Loop() {
		campo.SetInt(7)
	}
}

func BenchmarkStringDireta(b *testing.B) {
	d := &Dados{Nome: "go"}
	for b.Loop() {
		sinkStr = d.Nome
	}
}

// Interface() devolve um any: para tipos que não cabem em uma palavra, pode alocar.
func BenchmarkStringViaInterface(b *testing.B) {
	d := &Dados{Nome: "go"}
	campo := reflect.ValueOf(d).Elem().Field(2)
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = campo.Interface().(string)
	}
}

func BenchmarkStringViaString(b *testing.B) {
	d := &Dados{Nome: "go"}
	campo := reflect.ValueOf(d).Elem().Field(2)
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = campo.String()
	}
}

func BenchmarkChamarMetodoDireto(b *testing.B) {
	c := Calc{}
	for b.Loop() {
		sinkInt = c.Dobro(21)
	}
}

func BenchmarkChamarMetodoReflexao(b *testing.B) {
	m := reflect.ValueOf(Calc{}).MethodByName("Dobro")
	args := []reflect.Value{reflect.ValueOf(21)}
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = int(m.Call(args)[0].Int())
	}
}

// Calc tem um método para o benchmark.
type Calc struct{}

// Dobro dobra o número.
func (Calc) Dobro(n int) int { return n * 2 }
