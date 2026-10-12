package main

import (
	"cmp"
	"container/heap"
	"fmt"
	"math/rand/v2"
	"slices"
	"sort"
	"testing"
)

// ---- sort.Sort é um Template Method: o algoritmo é fixo e chama três
// passos que a SUA coleção implementa (Len, Less, Swap).

type Pessoa struct {
	Nome  string
	Idade int
}

type PorIdade []Pessoa

func (p PorIdade) Len() int           { return len(p) }
func (p PorIdade) Less(i, j int) bool { return p[i].Idade < p[j].Idade }
func (p PorIdade) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Espiao embrulha uma sort.Interface e conta as chamadas aos passos.
type Espiao struct {
	sort.Interface
	Less_, Swap_ int
}

func (e *Espiao) Less(i, j int) bool { e.Less_++; return e.Interface.Less(i, j) }
func (e *Espiao) Swap(i, j int)      { e.Swap_++; e.Interface.Swap(i, j) }

// ---- container/heap: outro molde. Push, Pop, Fix e Init chamam
// Len/Less/Swap e os dois passos Push/Pop da SUA coleção.

type Tarefa struct {
	Nome       string
	Prioridade int
}

type Fila []Tarefa

func (f Fila) Len() int           { return len(f) }
func (f Fila) Less(i, j int) bool { return f[i].Prioridade > f[j].Prioridade }
func (f Fila) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f *Fila) Push(x any)        { *f = append(*f, x.(Tarefa)) }
func (f *Fila) Pop() any {
	v := (*f)[len(*f)-1]
	*f = (*f)[:len(*f)-1]
	return v
}

func main() {
	fmt.Println("== sort.Sort: o molde chama os nossos passos ==")
	r := rand.New(rand.NewPCG(1, 2))
	for _, n := range []int{10, 100, 1000, 10000} {
		ps := make(PorIdade, n)
		for i := range ps {
			ps[i] = Pessoa{fmt.Sprint(i), r.IntN(100000)}
		}
		e := &Espiao{Interface: ps}
		sort.Sort(e)
		fmt.Printf("n=%-6d Less chamado %7d vezes, Swap %6d vezes (ordenado: %v)\n", n, e.Less_, e.Swap_, sort.IsSorted(ps))
	}

	fmt.Println("== container/heap: fila de prioridade ==")
	f := &Fila{}
	heap.Init(f)
	for _, t := range []Tarefa{{"backup", 2}, {"deploy", 9}, {"relatório", 5}, {"limpeza", 1}, {"alerta", 10}} {
		heap.Push(f, t)
	}
	for f.Len() > 0 {
		t := heap.Pop(f).(Tarefa)
		fmt.Printf("  %-10s prioridade %d\n", t.Nome, t.Prioridade)
	}

	fmt.Println("== Template Method (sort.Sort) x Strategy (slices.SortFunc), 100000 pessoas ==")
	base := make([]Pessoa, 100000)
	for i := range base {
		base[i] = Pessoa{fmt.Sprint(i), r.IntN(1000000)}
	}
	bench := func(f func(ps []Pessoa)) testing.BenchmarkResult {
		return testing.Benchmark(func(b *testing.B) {
			ps := make([]Pessoa, len(base))
			for b.Loop() {
				b.StopTimer()
				copy(ps, base)
				b.StartTimer()
				f(ps)
			}
		})
	}
	rt := bench(func(ps []Pessoa) { sort.Sort(PorIdade(ps)) })
	rs := bench(func(ps []Pessoa) {
		slices.SortFunc(ps, func(a, b Pessoa) int { return cmp.Compare(a.Idade, b.Idade) })
	})
	fmt.Printf("sort.Sort(PorIdade)   %7.2f ms/op  %d allocs\n", float64(rt.NsPerOp())/1e6, rt.AllocsPerOp())
	fmt.Printf("slices.SortFunc(cmp)  %7.2f ms/op  %d allocs  (%.1fx mais rápido)\n",
		float64(rs.NsPerOp())/1e6, rs.AllocsPerOp(), float64(rt.NsPerOp())/float64(rs.NsPerOp()))
}
