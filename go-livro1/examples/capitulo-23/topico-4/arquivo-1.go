package main

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// SemAcentos remove marcas diacriticas: decompoe (NFD), apaga as marcas (Mn), recompoe (NFC).
var semAcentos = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func SemAcentos(s string) string {
	r, _, err := transform.String(semAcentos, s)
	if err != nil {
		return s
	}
	return r
}

var naoAlfanum = regexp.MustCompile(`[^a-z0-9]+`)

// Slug gera um identificador de URL: minusculas, sem acentos, hifens.
func Slug(s string) string {
	s = strings.ToLower(SemAcentos(s))
	s = naoAlfanum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func main() {
	nfc := "café"
	nfd := "café"
	fmt.Println("iguais por bytes:      ", nfc == nfd)
	fmt.Println("iguais apos NFC:       ", norm.NFC.String(nfc) == norm.NFC.String(nfd))
	fmt.Printf("bytes NFC=%d NFD=%d\n", len(norm.NFC.String(nfd)), len(norm.NFD.String(nfc)))

	// o mesmo login digitado em duas formas
	usuarios := map[string]bool{norm.NFC.String("josé"): true}
	digitado := "josé" // vindo de um teclado/sistema que usa NFD
	fmt.Println("busca sem normalizar:  ", usuarios[digitado])
	fmt.Println("busca normalizando:    ", usuarios[norm.NFC.String(digitado)])

	fmt.Println()
	fmt.Printf("SemAcentos:  %q\n", SemAcentos("Ação, Coração & Pão de Açúcar - Ñandú"))
	fmt.Printf("Slug:        %q\n", Slug("  Introdução à Programação em Go!! (2ª edição) "))

	// ordenar em portugues
	palavras := []string{"zebra", "Árvore", "abelha", "ação", "Zorro", "avião", "casa", "Casa"}
	porBytes := slices.Clone(palavras)
	slices.Sort(porBytes)
	fmt.Println("\nordem por bytes:     ", porBytes)

	cl := collate.New(language.BrazilianPortuguese)
	porLocale := slices.Clone(palavras)
	cl.SortStrings(porLocale)
	fmt.Println("ordem pt-BR (collate):", porLocale)

	// sensivel a acento? o collate tem opcoes
	semAcc := collate.New(language.BrazilianPortuguese, collate.Loose)
	fmt.Println("Compare(\"ação\",\"acao\") padrao:", cl.CompareString("ação", "acao"), "| Loose:", semAcc.CompareString("ação", "acao"))

	// maiusculas e titulos respeitando o idioma
	fmt.Println()
	fmt.Println("strings.ToUpper(\"istanbul\"): ", strings.ToUpper("istanbul"))
	fmt.Println("cases turco Upper:           ", cases.Upper(language.Turkish).String("istanbul"))
	fmt.Println("cases Title pt-BR:           ", cases.Title(language.BrazilianPortuguese).String("a guerra dos mundos"))
	fmt.Println("cases Fold (comparar):       ", cases.Fold().String("STRASSE") == cases.Fold().String("straße"))
}
