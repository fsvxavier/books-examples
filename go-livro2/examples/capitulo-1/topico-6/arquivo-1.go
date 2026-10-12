package main

import (
	"fmt"
	"strings"
)

// ---- Tentativa de "heranca" por embutimento ----

type Base struct{}

func (b Base) Passo() string { return "passo da BASE" }

// Executar chama b.Passo(): e SEMPRE o Passo de Base. Em Java seria virtual.
func (b Base) Executar() string { return "executando: " + b.Passo() }

type Derivado struct{ Base }

func (d Derivado) Passo() string { return "passo do DERIVADO" }

// ---- Template Method em Go: a "parte variavel" e passada explicitamente ----

type Passos interface {
	Abrir() string
	Processar() string
	Fechar() string
}

// Rodar e o metodo-molde: fixa a ordem, delega os passos.
func Rodar(p Passos) string {
	return strings.Join([]string{p.Abrir(), p.Processar(), p.Fechar()}, " > ")
}

type ExportaCSV struct{}

func (ExportaCSV) Abrir() string     { return "abre csv" }
func (ExportaCSV) Processar() string { return "grava linhas" }
func (ExportaCSV) Fechar() string    { return "fecha csv" }

// Passos padrao reaproveitaveis por embutimento + sobrescrita de UM passo.
type PassosPadrao struct{}

func (PassosPadrao) Abrir() string     { return "abre" }
func (PassosPadrao) Processar() string { return "(nada)" }
func (PassosPadrao) Fechar() string    { return "fecha" }

type ExportaJSON struct{ PassosPadrao }

func (ExportaJSON) Processar() string { return "grava objetos" }

// ---- Embutir uma INTERFACE: decorar mudando so um metodo ----

type Logger interface {
	Info(msg string) string
	Erro(msg string) string
}

type logPadrao struct{}

func (logPadrao) Info(m string) string { return "INFO " + m }
func (logPadrao) Erro(m string) string { return "ERRO " + m }

// SemErro herda Info de graca e troca so o Erro.
type SemErro struct{ Logger }

func (SemErro) Erro(m string) string { return "(erro engolido)" }

func main() {
	d := Derivado{}
	fmt.Println("d.Passo():    ", d.Passo())
	fmt.Println("d.Executar(): ", d.Executar(), "<- usou o Passo da Base, nao o do Derivado")

	fmt.Println()
	fmt.Println("Rodar(CSV): ", Rodar(ExportaCSV{}))
	fmt.Println("Rodar(JSON):", Rodar(ExportaJSON{}))

	fmt.Println()
	var l Logger = SemErro{logPadrao{}}
	fmt.Println(l.Info("tudo bem"))
	fmt.Println(l.Erro("falhou"))
}
