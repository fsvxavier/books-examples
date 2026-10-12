package main

import (
	"fmt"
	"strings"
)

// Grade é um tabuleiro retangular do Jogo da Vida. true significa célula viva.
type Grade [][]bool

func NovaGrade(linhas, colunas int) Grade {
	g := make(Grade, linhas)
	for i := range g {
		g[i] = make([]bool, colunas)
	}
	return g
}

// FromTexto monta uma grade a partir de linhas de texto: '#' é viva, '.' é morta.
func FromTexto(linhas ...string) Grade {
	g := NovaGrade(len(linhas), len(linhas[0]))
	for i, l := range linhas {
		for j, c := range l {
			g[i][j] = c == '#'
		}
	}
	return g
}

func (g Grade) String() string {
	var sb strings.Builder
	for _, linha := range g {
		for _, viva := range linha {
			if viva {
				sb.WriteByte('#')
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// vizinhos conta as células vivas ao redor, com as bordas "dando a volta" (toro).
func (g Grade) vizinhos(i, j int) int {
	n := 0
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ii := (i + di + len(g)) % len(g)
			jj := (j + dj + len(g[0])) % len(g[0])
			if g[ii][jj] {
				n++
			}
		}
	}
	return n
}

// Proxima calcula a geração seguinte. Precisa de uma grade nova: atualizar no lugar
// misturaria células da geração atual com as da seguinte.
func (g Grade) Proxima() Grade {
	nova := NovaGrade(len(g), len(g[0]))
	for i := range g {
		for j := range g[i] {
			v := g.vizinhos(i, j)
			nova[i][j] = v == 3 || (g[i][j] && v == 2)
		}
	}
	return nova
}

func main() {
	g := FromTexto(
		".#....",
		"..#...",
		"###...",
		"......",
		"......",
		"......",
	)
	for geracao := range 3 {
		fmt.Printf("geração %d:\n%v\n", geracao, g)
		g = g.Proxima()
	}
}
