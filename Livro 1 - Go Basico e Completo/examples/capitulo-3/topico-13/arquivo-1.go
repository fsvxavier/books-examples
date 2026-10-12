package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Venda struct {
	Produto   string
	Categoria string
	Valor     int64 // centavos
}

type Resumo struct {
	Quantidade int
	Total      int64
}

func reais(c int64) string {
	return fmt.Sprintf("R$ %d,%02d", c/100, c%100)
}

func main() {
	vendas := []Venda{
		{"Caneca", "casa", 2990},
		{"Camiseta", "roupa", 5990},
		{"Adesivo", "papelaria", 350},
		{"Boné", "roupa", 4500},
		{"Caderno", "papelaria", 1890},
		{"Caneca", "casa", 2990},
		{"Vaso", "casa", 7900},
	}

	porCategoria := make(map[string]Resumo)
	for _, v := range vendas {
		r := porCategoria[v.Categoria]
		r.Quantidade++
		r.Total += v.Valor
		porCategoria[v.Categoria] = r
	}

	var totalGeral int64
	fmt.Printf("%-12s %5s %12s %8s\n", "categoria", "qtd", "total", "ticket")
	fmt.Println(strings.Repeat("-", 40))
	for _, cat := range slices.Sorted(maps.Keys(porCategoria)) {
		r := porCategoria[cat]
		fmt.Printf("%-12s %5d %12s %8s\n", cat, r.Quantidade, reais(r.Total), reais(r.Total/int64(r.Quantidade)))
		totalGeral += r.Total
	}
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("%-12s %5d %12s\n", "total", len(vendas), reais(totalGeral))

	contagem := make(map[string]int)
	for _, v := range vendas {
		contagem[v.Produto]++
	}
	for _, p := range slices.Sorted(maps.Keys(contagem)) {
		if contagem[p] > 1 {
			fmt.Printf("%s vendido %d vezes\n", p, contagem[p])
		}
	}
}
