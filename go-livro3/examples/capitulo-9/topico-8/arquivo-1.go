package main

import "fmt"

// grafo guarda as arestas de saída em CSR: os vizinhos do nó v ficam em
// idx[ptr[v]:ptr[v+1]]. Não há um ponteiro por aresta, só duas fatias.
type grafo struct {
	ptr []int32
	idx []int32
}

type aresta struct{ de, para int32 }

func construir(n int, arestas []aresta) grafo {
	ptr := make([]int32, n+1)
	for _, a := range arestas {
		ptr[a.de+1]++ // conta o grau de saída de cada nó
	}
	for i := 0; i < n; i++ {
		ptr[i+1] += ptr[i] // soma em prefixo: ptr[v] vira o início da lista de v
	}
	idx := make([]int32, len(arestas))
	proximo := append([]int32(nil), ptr[:n]...)
	for _, a := range arestas {
		idx[proximo[a.de]] = a.para
		proximo[a.de]++
	}
	return grafo{ptr, idx}
}

func (g grafo) saida(v int32) []int32 { return g.idx[g.ptr[v]:g.ptr[v+1]] }

// largura devolve a distância, em arestas, de origem até cada nó (-1 se inalcançável).
func (g grafo) largura(origem int32) []int {
	dist := make([]int, len(g.ptr)-1)
	for i := range dist {
		dist[i] = -1
	}
	dist[origem] = 0
	fila := []int32{origem}
	for len(fila) > 0 {
		v := fila[0]
		fila = fila[1:]
		for _, w := range g.saida(v) {
			if dist[w] < 0 {
				dist[w] = dist[v] + 1
				fila = append(fila, w)
			}
		}
	}
	return dist
}

func main() {
	g := construir(6, []aresta{{0, 1}, {0, 2}, {1, 3}, {2, 3}, {3, 4}})
	fmt.Println("ptr:", g.ptr)
	fmt.Println("idx:", g.idx)
	fmt.Println("saida(0):", g.saida(0), "saida(3):", g.saida(3), "saida(5):", g.saida(5))
	fmt.Println("distancias a partir de 0:", g.largura(0))
}
