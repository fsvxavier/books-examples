package main

import (
	"fmt"
	"slices"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

func main() {
	g := simple.NewDirectedGraph()
	pagar := func(de, para int64) {
		g.SetEdge(simple.Edge{F: simple.Node(de), T: simple.Node(para)})
	}
	// Contas 1, 2 e 3 formam um consórcio legítimo: cada uma paga à seguinte.
	pagar(1, 2)
	pagar(2, 3)
	pagar(3, 1)
	// Contas 4, 5 e 6 são uma cadeia comum, sem ciclo. A conta 2 também paga à 4.
	pagar(2, 4)
	pagar(4, 5)
	pagar(5, 6)

	// A ordem em que o Tarjan devolve os componentes não é garantida; ordenamos para imprimir.
	var comps [][]int64
	for _, comp := range topo.TarjanSCC(g) {
		comps = append(comps, ids(comp))
	}
	slices.SortFunc(comps, func(a, b []int64) int { return int(a[0] - b[0]) })
	for _, c := range comps {
		fmt.Println("componente fortemente conexo:", c)
	}

	// O simple.DirectedGraph não aceita laço próprio e entra em pânico.
	func() {
		defer func() { fmt.Println("SetEdge(7 -> 7):", recover()) }()
		pagar(7, 7)
	}()
	fmt.Println("nó 99 existe?", g.Node(99) != nil)
}

func ids(nos []graph.Node) []int64 {
	r := make([]int64, len(nos))
	for i, n := range nos {
		r[i] = n.ID()
	}
	slices.Sort(r)
	return r
}
