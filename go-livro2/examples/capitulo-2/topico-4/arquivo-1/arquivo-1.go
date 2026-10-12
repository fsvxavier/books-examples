package main

import (
	"errors"
	"fmt"
)

type Leitor interface{ Ler() (string, error) }

type arquivo struct{ nome string }

func (a *arquivo) Ler() (string, error) { return "conteudo de " + a.nome, nil }

// Extra: metodo que NAO esta na interface.
func (a *arquivo) Tamanho() int { return len(a.nome) }

// ---- ARMADILHA 1: devolver um ponteiro nulo tipado dentro de uma interface ----

func AbrirRuim(nome string) (Leitor, error) {
	var a *arquivo // nil
	if nome != "" {
		a = &arquivo{nome: nome}
	}
	var err error
	if nome == "" {
		err = errors.New("nome vazio")
	}
	return a, err // BUG: quando a e nil, a interface NAO e nil
}

func AbrirBom(nome string) (Leitor, error) {
	if nome == "" {
		return nil, errors.New("nome vazio") // nil literal: interface nula de verdade
	}
	return &arquivo{nome: nome}, nil
}

// ---- ARMADILHA 2: a interface esconde os metodos extras ----

// Quando so existe UMA implementacao, devolver o tipo concreto (exportado) e melhor.
func NovoArquivo(nome string) *arquivo { return &arquivo{nome: nome} }

func main() {
	demoBom()
	l, err := AbrirRuim("")
	fmt.Println("AbrirRuim(\"\"):  err =", err, "| l == nil?", l == nil, "<- a interface NAO e nil")
	if l != nil {
		defer func() { fmt.Println("  usar l: panic ->", recover()) }()
		l.Ler()
	}
}

func demoBom() {
	l, err := AbrirBom("")
	fmt.Println("AbrirBom(\"\"):   err =", err, "| l == nil?", l == nil)

	l, _ = AbrirBom("a.txt")
	// l.Tamanho() nao compila: a interface Leitor so tem Ler.
	if t, ok := l.(interface{ Tamanho() int }); ok {
		fmt.Println("via asserção de tipo, Tamanho =", t.Tamanho())
	}
	fmt.Println("tipo concreto devolvido, Tamanho =", NovoArquivo("a.txt").Tamanho())
}
