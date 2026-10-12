package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	// Pipeline sobre uma sequencia INFINITA: os 5 primeiros quadrados de numeros pares.
	var origem int
	pares := Filtrar(Contar(Naturais(), &origem), func(n int) bool { return n%2 == 0 })
	quadrados := Mapear(pares, func(n int) int { return n * n })
	primeiros := slices.Collect(Limitar(quadrados, 5))
	fmt.Println("primeiros 5 quadrados de pares:", primeiros)
	fmt.Println("elementos que a origem infinita chegou a produzir:", origem)

	// Nada acontece antes de alguem consumir:
	var chamadas int
	montado := Mapear(slices.Values([]int{1, 2, 3}), func(n int) int { chamadas++; return n * 10 })
	fmt.Println("pipeline montado, chamadas ate agora:", chamadas)
	_ = slices.Collect(montado)
	fmt.Println("depois de consumir:", chamadas)

	// Zip termina com a menor.
	nomes := slices.Values([]string{"ana", "bia", "caio"})
	for nome, n := range Zip(nomes, Naturais()) {
		fmt.Printf("%d=%s ", n, nome)
	}
	fmt.Println()

	// Os adaptadores da biblioteca padrao: maps.All, slices.Sorted, slices.Chunk, strings.SplitSeq.
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	fmt.Println("chaves ordenadas:", slices.Sorted(maps.Keys(m)))
	for pedaco := range slices.Chunk([]int{1, 2, 3, 4, 5}, 2) {
		fmt.Print(pedaco, " ")
	}
	fmt.Println()
	for parte := range strings.SplitSeq("a,b,c", ",") {
		fmt.Print(parte, ";")
	}
	fmt.Println()

	soma := Reduzir(Limitar(Naturais(), 100), 0, func(acc, v int) int { return acc + v })
	fmt.Println("soma de 1 a 100:", soma)
}
