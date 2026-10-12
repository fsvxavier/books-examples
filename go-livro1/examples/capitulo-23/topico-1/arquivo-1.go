package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

func medidas(rotulo, s string) {
	fmt.Printf("%-12s %-10q bytes=%-2d runes=%-2d graphemes=%-2d\n",
		rotulo, s, len(s), utf8.RuneCountInString(s), uniseg.GraphemeClusterCount(s))
}

// TruncarGrafemas corta em n caracteres PERCEBIDOS, sem partir emoji nem acentos.
func TruncarGrafemas(s string, n int) string {
	var b strings.Builder
	g := uniseg.NewGraphemes(s)
	for i := 0; i < n && g.Next(); i++ {
		b.WriteString(g.Str())
	}
	return b.String()
}

func main() {
	fmt.Println("versao do Unicode no Go:", unicode.Version)
	fmt.Println()

	medidas("ascii", "go")
	medidas("acento NFC", "café")  // e com acento: 1 ponto de codigo
	medidas("acento NFD", "café") // e + acento combinante: 2 pontos de codigo
	medidas("emoji", "😀")
	medidas("familia", "👨‍👩‍👧") // 3 emojis unidos por ZWJ
	medidas("bandeira", "🇧🇷")   // 2 indicadores regionais
	medidas("polegar+tom", "👍🏽")
	fmt.Println("NFC == NFD (comparacao de bytes):", "café" == "café")

	// range sobre string devolve RUNAS, com o indice do byte
	for i, r := range "aé😀" {
		fmt.Printf("  indice=%d runa=%U %q (%d bytes)\n", i, r, r, utf8.RuneLen(r))
	}

	// bytes invalidos viram U+FFFD
	ruim := "ab\xffcd"
	fmt.Println("\nvalida UTF-8:", utf8.ValidString(ruim))
	for i, r := range ruim {
		fmt.Printf("  %d:%U", i, r)
	}
	fmt.Println()
	fmt.Printf("ToValidUTF8: %q\n", strings.ToValidUTF8(ruim, "?"))

	// inverter: por bytes quebra, por runas quebra o acento combinante, por grafemas nao
	s := "café 👨‍👩‍👧!"
	rs := []rune(s)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	fmt.Printf("\noriginal:       %s\n", s)
	fmt.Printf("invertido runas:%s   <- acento e familia embaralhados\n", string(rs))
	var gs []string
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		gs = append(gs, g.Str())
	}
	for i, j := 0, len(gs)-1; i < j; i, j = i+1, j-1 {
		gs[i], gs[j] = gs[j], gs[i]
	}
	fmt.Printf("invertido graf.:%s\n", strings.Join(gs, ""))

	// truncar
	fmt.Printf("\ncortar 7 bytes:     %q\n", "olá, 😀 mundo"[:7])
	fmt.Printf("cortar 7 grafemas:  %q\n", TruncarGrafemas("olá, 😀 mundo", 7))
}
