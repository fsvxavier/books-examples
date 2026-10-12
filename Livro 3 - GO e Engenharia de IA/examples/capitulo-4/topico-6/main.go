package main

import (
	"fmt"
	"os"
	"sort"
)

// comandos reúne os subcomandos do laboratório; cada arquivo registra o seu em init.
var comandos = map[string]func([]string) error{}

func registra(nome string, f func([]string) error) { comandos[nome] = f }

func main() {
	if len(os.Args) < 2 || comandos[os.Args[1]] == nil {
		var nomes []string
		for n := range comandos {
			nomes = append(nomes, n)
		}
		sort.Strings(nomes)
		fmt.Fprintln(os.Stderr, "uso: lab <comando> [argumentos]; comandos:", nomes)
		os.Exit(2)
	}
	if err := comandos[os.Args[1]](os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

// cortaTxt devolve s limitada a n caracteres, trocando o resto por reticências.
func cortaTxt(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

// encurta reduz um valor JSON decodificado para caber em uma linha de log:
// textos longos são cortados e listas ficam só com os primeiros itens.
func encurta(v any) any {
	switch x := v.(type) {
	case string:
		return cortaTxt(x, 60)
	case []any:
		const max = 3
		out := make([]any, 0, max+1)
		for i, e := range x {
			if i == max {
				out = append(out, fmt.Sprintf("... (+%d)", len(x)-max))
				break
			}
			out = append(out, encurta(e))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = encurta(e)
		}
		return out
	}
	return v
}
