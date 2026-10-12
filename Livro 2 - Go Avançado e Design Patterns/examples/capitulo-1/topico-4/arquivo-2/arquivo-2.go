package main

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// ---- Interface gorda: o "repositorio de tudo" ----

type RepositorioGordo interface {
	Criar(nome string) (int, error)
	Buscar(id int) (string, error)
	Atualizar(id int, nome string) error
	Remover(id int) error
	Listar() ([]string, error)
	Contar() (int, error)
	Exportar(w io.Writer) error
	Importar(r io.Reader) error
}

// Quem so precisa buscar nome define a SUA interface, no ponto de uso.
type BuscadorDeNome interface {
	Buscar(id int) (string, error)
}

func Saudacao(b BuscadorDeNome, id int) string {
	nome, err := b.Buscar(id)
	if err != nil {
		return "ola, visitante"
	}
	return "ola, " + nome
}

// O fake do teste implementa UM metodo.
type buscadorFake map[int]string

func (f buscadorFake) Buscar(id int) (string, error) {
	if n, ok := f[id]; ok {
		return n, nil
	}
	return "", fmt.Errorf("id %d nao existe", id)
}

// Verificacao em tempo de compilacao: o tipo concreto satisfaz a interface pequena.
var _ BuscadorDeNome = buscadorFake(nil)

func main() {
	fmt.Println("metodos de RepositorioGordo:", reflect.TypeFor[RepositorioGordo]().NumMethod())
	fmt.Println("metodos de BuscadorDeNome:  ", reflect.TypeFor[BuscadorDeNome]().NumMethod())
	fmt.Println(Saudacao(buscadorFake{1: "Ana"}, 1))
	fmt.Println(Saudacao(buscadorFake{1: "Ana"}, 9))

	// a biblioteca padrao: interfaces de UM metodo que se compoem
	fmt.Println()
	fmt.Println("io.Reader:     ", reflect.TypeFor[io.Reader]().NumMethod(), "metodo")
	fmt.Println("io.Writer:     ", reflect.TypeFor[io.Writer]().NumMethod(), "metodo")
	fmt.Println("io.ReadWriter: ", reflect.TypeFor[io.ReadWriter]().NumMethod(), "metodos (Reader + Writer embutidos)")
	fmt.Println("io.ReadCloser: ", reflect.TypeFor[io.ReadCloser]().NumMethod(), "metodos")

	// uma funcao que aceita io.Reader serve a arquivo, rede, string, buffer, gzip...
	contar := func(r io.Reader) int { n, _ := io.Copy(io.Discard, r); return int(n) }
	fmt.Println("contar(strings.Reader):", contar(strings.NewReader("ola")))
	fmt.Println("contar(bytes.Buffer):  ", contar(bytes.NewBufferString("ola, mundo")))
}
