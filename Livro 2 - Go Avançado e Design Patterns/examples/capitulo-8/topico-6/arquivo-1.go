package main

import (
	"fmt"
	"runtime"
)

type Tipo struct {
	Nome    string // campos EXPORTADOS: qualquer um pode mudar o estado compartilhado
	Cor     string
	Textura []byte
}

type Arvore struct {
	X, Y float32
	Tipo *Tipo
}

type ArvoreCheia struct {
	X, Y    float32
	Nome    string
	Cor     string
	Textura []byte
}

func medir(f func() any) (any, uint64) {
	var a, d runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&a)
	v := f()
	runtime.GC()
	runtime.ReadMemStats(&d)
	return v, d.HeapAlloc - a.HeapAlloc
}

func tabela() {
	fmt.Println("Quanto o Flyweight economiza conforme o numero de arvores (3 tipos, textura de 1 KiB):")
	fmt.Printf("%10s %14s %14s %14s\n", "arvores", "sem (KiB)", "com (KiB)", "economia")
	for _, n := range []int{10, 100, 1_000, 10_000, 100_000} {
		v1, b1 := medir(func() any {
			fl := make([]ArvoreCheia, 0, n)
			for i := range n {
				fl = append(fl, ArvoreCheia{Nome: "n", Cor: "c", Textura: make([]byte, 1024)})
				_ = i
			}
			return fl
		})
		v2, b2 := medir(func() any {
			pool := []*Tipo{{"a", "x", make([]byte, 1024)}, {"b", "y", make([]byte, 1024)}, {"c", "z", make([]byte, 1024)}}
			fl := make([]Arvore, 0, n)
			for i := range n {
				fl = append(fl, Arvore{Tipo: pool[i%3]})
			}
			return struct {
				fl   []Arvore
				pool []*Tipo
			}{fl, pool}
		})
		fmt.Printf("%10d %14.1f %14.1f %14.1f\n", n, float64(b1)/1024, float64(b2)/1024, (float64(b1)-float64(b2))/1024)
		runtime.KeepAlive(v1)
		runtime.KeepAlive(v2)
	}
}

func bugMutacao() {
	fmt.Println("\nO bug classico: flyweight mutavel")
	carvalho := &Tipo{Nome: "carvalho", Cor: "verde"}
	floresta := []Arvore{{1, 1, carvalho}, {2, 2, carvalho}, {3, 3, carvalho}}

	// "Pintar de vermelho SO a arvore 0": o codigo altera o tipo, que e compartilhado.
	floresta[0].Tipo.Cor = "vermelho"
	for i, a := range floresta {
		fmt.Printf("  arvore %d: %s\n", i, a.Tipo.Cor)
	}
	fmt.Println("  (as tres ficaram vermelhas: a cor e estado EXTRINSECO e nao devia estar no flyweight)")
}

func main() {
	tabela()
	bugMutacao()
}
