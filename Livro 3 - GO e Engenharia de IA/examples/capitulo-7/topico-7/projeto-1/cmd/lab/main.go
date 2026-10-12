// Comando lab reúne os subcomandos do laboratório de features.
package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"time"
)

// inicio é o instante da primeira transação sintética.
var inicio = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

// transacao é uma compra de um cartão.
type transacao struct {
	Cartao string
	Quando time.Time
	Centav int64
}

// dados gera, com semente fixa, transações sintéticas de 500 cartões ao longo
// de 72 horas, ordenadas no tempo: os mesmos dados em toda execução.
func dados() []transacao {
	r := rand.New(rand.NewPCG(7, 21))
	var txs []transacao
	for c := range 500 {
		n := 20 + r.IntN(60)
		for range n {
			txs = append(txs, transacao{
				Cartao: fmt.Sprintf("c%05d", c),
				Quando: inicio.Add(time.Duration(r.Int64N(int64(72 * time.Hour)))),
				Centav: int64(500 + r.IntN(50000)),
			})
		}
	}
	slices.SortFunc(txs, func(a, b transacao) int { return a.Quando.Compare(b.Quando) })
	return txs
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "exportar" {
		fmt.Fprintln(os.Stderr, "uso: lab exportar [arquivo.csv]")
		os.Exit(2)
	}
	exportar()
}
