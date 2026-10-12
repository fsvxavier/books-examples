package main

import (
	"fmt"
	"math"
)

// Um passo de passagem de mensagens, à mão, em um grafo de 4 contas:
// 0 -> 1, 1 -> 2, 2 -> 0 (um ciclo) e 3 -> 2. Cada conta tem um único atributo.
func main() {
	x := []float64{1, 0, 0, 4}                          // a conta 3 tem atributo grande
	arestas := [][2]int{{0, 1}, {1, 2}, {2, 0}, {3, 2}} // {de, para}, em ordem fixa
	n := len(x)

	entradas := make([][]int, n)
	for _, a := range arestas {
		entradas[a[1]] = append(entradas[a[1]], a[0])
	}

	// Média dos vizinhos de entrada (o que a GraphSAGE usa): quem pagou à conta.
	fmt.Println("media dos vizinhos de entrada:")
	for v := 0; v < n; v++ {
		soma := 0.0
		for _, u := range entradas[v] {
			soma += x[u]
		}
		m := 0.0
		if len(entradas[v]) > 0 {
			m = soma / float64(len(entradas[v]))
		}
		fmt.Printf("  conta %d: entradas %v -> %.2f\n", v, entradas[v], m)
	}

	// Normalização simétrica da GCN: cada vizinho u contribui com x[u]/sqrt(gv*gu),
	// com os graus contando o laço próprio e as duas direções.
	viz := make([][]int, n)
	for v := 0; v < n; v++ {
		viz[v] = append(viz[v], v)
	}
	for _, a := range arestas {
		viz[a[0]] = append(viz[a[0]], a[1])
		viz[a[1]] = append(viz[a[1]], a[0])
	}
	fmt.Println("agregacao da GCN (A+Aᵀ+I normalizada):")
	for v := 0; v < n; v++ {
		soma := 0.0
		for _, u := range viz[v] {
			soma += x[u] / math.Sqrt(float64(len(viz[v]))*float64(len(viz[u])))
		}
		fmt.Printf("  conta %d: vizinhos %v -> %.3f\n", v, viz[v], soma)
	}
}
