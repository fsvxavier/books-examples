package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

const lado = 1000 // mapa de 1000 x 1000 = 1 milhao de celulas

// Tile e o estado INTRINSECO de um tipo de terreno: igual para todas as celulas do mesmo tipo.
type Tile struct {
	Nome   string
	Custo  int
	Sprite [64]byte
}

var tipos = [...]Tile{
	{Nome: "grama", Custo: 1},
	{Nome: "agua", Custo: 9},
	{Nome: "floresta", Custo: 3},
	{Nome: "montanha", Custo: 7},
}

func tipoDe(x, y int) int { return (x*7 + y*13 + x*y) % len(tipos) }

// 1) Copia do Tile em cada celula.
func mapaCopias() [][]Tile {
	m := make([][]Tile, lado)
	for y := range m {
		m[y] = make([]Tile, lado)
		for x := range m[y] {
			m[y][x] = tipos[tipoDe(x, y)]
		}
	}
	return m
}

// 2) Ponteiro para um Tile compartilhado (Flyweight classico).
func mapaPonteiros() [][]*Tile {
	m := make([][]*Tile, lado)
	for y := range m {
		m[y] = make([]*Tile, lado)
		for x := range m[y] {
			m[y][x] = &tipos[tipoDe(x, y)]
		}
	}
	return m
}

// 3) Indice de 1 byte em uma tabela de Tiles (o "flyweight" vira um indice).
type MapaIndices struct {
	celulas []uint8
	tabela  [len(tipos)]Tile
}

func mapaIndices() *MapaIndices {
	m := &MapaIndices{celulas: make([]uint8, lado*lado), tabela: tipos}
	for y := range lado {
		for x := range lado {
			m.celulas[y*lado+x] = uint8(tipoDe(x, y))
		}
	}
	return m
}

func (m *MapaIndices) Tile(x, y int) *Tile { return &m.tabela[m.celulas[y*lado+x]] }

func medir(f func() any) (any, uint64) {
	var a, d runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&a)
	v := f()
	runtime.GC()
	runtime.ReadMemStats(&d)
	return v, d.HeapAlloc - a.HeapAlloc
}

func somaCopias(m [][]Tile) int {
	t := 0
	for y := range m {
		for x := range m[y] {
			t += m[y][x].Custo
		}
	}
	return t
}

func somaPonteiros(m [][]*Tile) int {
	t := 0
	for y := range m {
		for x := range m[y] {
			t += m[y][x].Custo
		}
	}
	return t
}

func somaIndices(m *MapaIndices) int {
	t := 0
	for _, i := range m.celulas {
		t += m.tabela[i].Custo
	}
	return t
}

func relatorio() {
	v1, b1 := medir(func() any { return mapaCopias() })
	v2, b2 := medir(func() any { return mapaPonteiros() })
	v3, b3 := medir(func() any { return mapaIndices() })
	fmt.Printf("mapa de %dx%d celulas (%d tipos), Tile = %d bytes\n", lado, lado, len(tipos), unsafe.Sizeof(Tile{}))
	fmt.Printf("copia do Tile por celula: %7.1f MiB\n", float64(b1)/(1<<20))
	fmt.Printf("ponteiro para o Tile:     %7.1f MiB\n", float64(b2)/(1<<20))
	fmt.Printf("indice de 1 byte:         %7.1f MiB\n", float64(b3)/(1<<20))
	fmt.Println("mesmo resultado nas tres formas:",
		somaCopias(v1.([][]Tile)) == somaPonteiros(v2.([][]*Tile)) && somaPonteiros(v2.([][]*Tile)) == somaIndices(v3.(*MapaIndices)))
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
	runtime.KeepAlive(v3)
}

func main() {
	relatorio()
}
