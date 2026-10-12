// Command e08 faz um passeio pelos pacotes genéricos da biblioteca padrão.
package main

import (
	"cmp"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"unique"
)

type Pessoa struct {
	Nome  string
	Idade int
}

func main() {
	// --- slices ---
	nums := []int{5, 2, 8, 2, 9, 1}
	slices.Sort(nums)
	fmt.Println("Sort:", nums)
	i, achou := slices.BinarySearch(nums, 8)
	fmt.Println("BinarySearch(8):", i, achou)
	nums = slices.Compact(nums)
	fmt.Println("Compact:", nums, "| Contains(9):", slices.Contains(nums, 9), "| Index(8):", slices.Index(nums, 8))
	nums = slices.Insert(nums, 1, 100, 101)
	nums = slices.Delete(nums, 0, 1)
	fmt.Println("Insert/Delete:", nums, "| Max:", slices.Max(nums), "| Min:", slices.Min(nums))
	slices.Reverse(nums)
	fmt.Println("Reverse:", nums)

	pessoas := []Pessoa{{"Bia", 30}, {"Ana", 25}, {"Caio", 30}, {"Dani", 25}}
	slices.SortStableFunc(pessoas, func(a, b Pessoa) int {
		return cmp.Or(cmp.Compare(b.Idade, a.Idade), strings.Compare(a.Nome, b.Nome)) // idade desc, nome asc
	})
	fmt.Println("SortStableFunc + cmp.Or:", pessoas)
	for bloco := range slices.Chunk([]int{1, 2, 3, 4, 5}, 2) {
		fmt.Print(" bloco=", bloco)
	}
	fmt.Println()
	fmt.Println("Concat/Repeat:", slices.Concat([]int{1}, []int{2, 3}), slices.Repeat([]string{"ab"}, 3))

	// --- maps ---
	estoque := map[string]int{"maca": 3, "pera": 0, "uva": 7}
	chaves := slices.Sorted(maps.Keys(estoque)) // maps.Keys devolve um iterador; Sorted o consome
	fmt.Println("Keys ordenadas:", chaves, "| soma dos Values:", func() (s int) {
		for v := range maps.Values(estoque) {
			s += v
		}
		return
	}())
	copia := maps.Clone(estoque)
	maps.DeleteFunc(copia, func(_ string, q int) bool { return q == 0 })
	fmt.Println("DeleteFunc (sem estoque zero):", len(copia), "itens | Equal com o original:", maps.Equal(copia, estoque))

	// --- cmp, min, max, clear ---
	fmt.Println("cmp.Compare:", cmp.Compare(1, 2), cmp.Compare("b", "a"), cmp.Compare(2.0, 2.0), "| min/max:", min(3, 1, 2), max(2.5, 1))
	clear(copia)
	fmt.Println("clear(map):", len(copia))

	// --- iteradores ---
	fmt.Println("Pull:", func() []int {
		prox, parar := iter.Pull(slices.Values([]int{10, 20, 30}))
		defer parar()
		var saida []int
		for {
			v, ok := prox()
			if !ok {
				return saida
			}
			saida = append(saida, v)
		}
	}())

	// --- sync/atomic e unique ---
	var conf atomic.Pointer[map[string]string]
	conf.Store(&map[string]string{"modo": "teste"})
	fmt.Println("atomic.Pointer[T]:", (*conf.Load())["modo"])

	calc := sync.OnceValue(func() int { fmt.Println("  (calculando uma vez)"); return 42 })
	fmt.Println("OnceValue:", calc(), calc())

	h1, h2 := unique.Make("go"), unique.Make("go")
	fmt.Println("unique.Handle: iguais =", h1 == h2, "| valor =", h1.Value())
}
