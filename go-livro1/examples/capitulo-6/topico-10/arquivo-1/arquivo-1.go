package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
)

// Indice é um índice invertido: para cada palavra, o conjunto de documentos onde ela aparece.
type Indice struct {
	docs     []string
	palavras map[string]map[int]struct{}
}

func NovoIndice() *Indice {
	return &Indice{palavras: make(map[string]map[int]struct{})}
}

// normalizar quebra o texto em palavras minúsculas, ignorando pontuação.
func normalizar(texto string) []string {
	return strings.FieldsFunc(strings.ToLower(texto), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// Adicionar indexa um documento e devolve o identificador dele.
func (ix *Indice) Adicionar(texto string) int {
	id := len(ix.docs)
	ix.docs = append(ix.docs, texto)
	for _, p := range normalizar(texto) {
		if ix.palavras[p] == nil {
			ix.palavras[p] = make(map[int]struct{})
		}
		ix.palavras[p][id] = struct{}{}
	}
	return id
}

// Buscar devolve os documentos que contêm TODAS as palavras da consulta (interseção).
func (ix *Indice) Buscar(consulta string) []int {
	termos := normalizar(consulta)
	if len(termos) == 0 {
		return nil
	}

	resultado := maps.Clone(ix.palavras[termos[0]])
	for _, t := range termos[1:] {
		for id := range resultado {
			if _, ok := ix.palavras[t][id]; !ok {
				delete(resultado, id)
			}
		}
	}
	return slices.Sorted(maps.Keys(resultado))
}

func main() {
	ix := NovoIndice()
	ix.Adicionar("Go é uma linguagem compilada e concorrente.")
	ix.Adicionar("Python é uma linguagem interpretada.")
	ix.Adicionar("Go usa goroutines para concorrência; Python usa threads.")
	ix.Adicionar("Rust é uma linguagem compilada, sem coletor de lixo.")

	for _, q := range []string{"linguagem", "compilada", "go", "linguagem compilada", "Python threads", "java", ""} {
		ids := ix.Buscar(q)
		fmt.Printf("%-22q -> %v\n", q, ids)
	}
	fmt.Println(len(ix.palavras), "palavras distintas")
}
