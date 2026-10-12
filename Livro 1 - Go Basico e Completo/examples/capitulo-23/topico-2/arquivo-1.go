package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main() {
	// Cut: separar em duas partes (substitui Index + fatiamento)
	chave, valor, ok := strings.Cut("host=localhost:5432", "=")
	fmt.Printf("Cut:            %q %q %v\n", chave, valor, ok)
	caminho, ext, _ := strings.CutLast("relatorio.final.v2.csv", ".") // Go 1.27
	fmt.Printf("CutLast:        %q %q\n", caminho, ext)
	rest, ok := strings.CutPrefix("Bearer abc123", "Bearer ")
	fmt.Printf("CutPrefix:      %q %v\n", rest, ok)
	nome, ok := strings.CutSuffix("foto.png", ".png")
	fmt.Printf("CutSuffix:      %q %v\n", nome, ok)

	// Split e as pegadinhas
	fmt.Printf("Split vazio:    %q (len=%d)\n", strings.Split("", ","), len(strings.Split("", ",")))
	fmt.Printf("Split normal:   %q\n", strings.Split("a,b,,c", ","))
	fmt.Printf("SplitN:         %q\n", strings.SplitN("a=b=c", "=", 2))
	fmt.Printf("Fields:         %q\n", strings.Fields("  um   dois\tte\nres  "))
	fmt.Printf("FieldsFunc:     %q\n", strings.FieldsFunc("um;dois, tres", func(r rune) bool { return r == ';' || r == ',' || unicode.IsSpace(r) }))

	// iteradores (Go 1.24): sem criar o slice inteiro
	var partes []string
	for p := range strings.SplitSeq("2026-10-04", "-") {
		partes = append(partes, p)
	}
	fmt.Printf("SplitSeq:       %q\n", partes)
	n := 0
	for linha := range strings.Lines("linha 1\nlinha 2\nultima sem quebra") {
		n++
		fmt.Printf("  Lines %d: %q\n", n, linha)
	}
	for p := range strings.FieldsSeq("a  b   c") {
		fmt.Print("  [", p, "]")
	}
	fmt.Println()

	// Trim: cuidado com o CONJUNTO de caracteres
	fmt.Printf("Trim:           %q\n", strings.Trim("xxhixyx", "xy"))
	fmt.Printf("TrimPrefix:     %q (so o prefixo exato)\n", strings.TrimPrefix("xxhi", "x"))
	fmt.Printf("TrimSpace:      %q\n", strings.TrimSpace("\t oi \n"))
	r, _ := strings.CutPrefix("abba", "ab")
	fmt.Printf("TrimLeft:       %q (conjunto {a,b})  CutPrefix: %q (texto exato)\n", strings.TrimLeft("abba", "ab"), r)

	// maiusculas, minusculas, comparacao
	fmt.Printf("ToUpper:        %q %q\n", strings.ToUpper("olá, ação"), strings.ToUpper("straße"))
	fmt.Printf("EqualFold:      %v %v\n", strings.EqualFold("Go", "GO"), strings.EqualFold("straße", "STRASSE"))
	fmt.Printf("Contains/Count: %v %d %d\n", strings.Contains("banana", "nan"), strings.Count("banana", "an"), strings.Count("banana", ""))
	fmt.Printf("Index:          %d %d\n", strings.Index("banana", "an"), strings.LastIndex("banana", "an"))
	fmt.Printf("ContainsAny:    %v   ContainsRune: %v\n", strings.ContainsAny("go", "xyzo"), strings.ContainsRune("go", 'g'))

	// Replacer: varias trocas em uma passada
	rp := strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	fmt.Printf("Replacer:       %q\n", rp.Replace("a < b && c > d"))
	fmt.Printf("Replace n=1:    %q  ReplaceAll: %q\n", strings.Replace("a-b-c", "-", "+", 1), strings.ReplaceAll("a-b-c", "-", "+"))

	// Map, Repeat, Join
	semDigitos := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return -1 // remove
		}
		return r
	}, "a1b22c333")
	fmt.Printf("Map:            %q\n", semDigitos)
	fmt.Printf("Repeat/Join:    %q %q\n", strings.Repeat("ab", 3), strings.Join([]string{"x", "y", "z"}, "/"))
}
