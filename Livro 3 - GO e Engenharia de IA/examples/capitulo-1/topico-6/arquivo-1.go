// Reproduz o defeito do rascunho antigo do capítulo: várias goroutines
// gravando no MESMO map sem trava.
package main

import (
	"fmt"
	"sync"
)

func buscarEmTodas(nomes []string) map[string][]string {
	resultados := make(map[string][]string)
	var wg sync.WaitGroup
	for _, nome := range nomes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resultados[nome] = []string{"doc1", "doc2"} // BUG: escrita concorrente em map
		}()
	}
	wg.Wait()
	return resultados
}

func main() {
	nomes := make([]string, 64)
	for i := range nomes {
		nomes[i] = fmt.Sprintf("loja%02d", i)
	}
	r := buscarEmTodas(nomes)
	fmt.Println("fontes com resultado:", len(r), "de", len(nomes))
}
