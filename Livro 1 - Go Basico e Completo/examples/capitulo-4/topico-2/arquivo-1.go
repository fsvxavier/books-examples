package main

import "fmt"

type Ponto struct{ X, Y int }

func moverCopia(p Ponto)       { p.X += 100 }
func moverPonteiro(p *Ponto)   { p.X += 100 }
func dobrarArray(a [3]int)     { a[0] *= 2 }
func dobrarSlice(s []int)      { s[0] *= 2 }
func crescerSlice(s []int)     { s = append(s, 99) }
func crescerPonteiro(s *[]int) { *s = append(*s, 99) }
func marcarMapa(m map[string]int) {
	m["novo"] = 1
}
func reatribuir(n int) { n = 50 }

func main() {
	p := Ponto{1, 2}
	moverCopia(p)
	fmt.Println("após cópia:", p)
	moverPonteiro(&p)
	fmt.Println("após ponteiro:", p)

	a := [3]int{1, 2, 3}
	dobrarArray(a)
	fmt.Println("array:", a)

	s := []int{1, 2, 3}
	dobrarSlice(s)
	fmt.Println("slice elementos:", s)
	crescerSlice(s)
	fmt.Println("slice após append na cópia:", s, len(s))
	crescerPonteiro(&s)
	fmt.Println("slice via ponteiro:", s, len(s))

	m := map[string]int{}
	marcarMapa(m)
	fmt.Println("mapa:", m)

	n := 10
	reatribuir(n)
	fmt.Println(n)
}
