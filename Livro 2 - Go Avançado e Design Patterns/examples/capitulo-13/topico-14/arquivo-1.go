package main

import (
	"fmt"
	"testing"
)

type No interface {
	Aceitar(Visitante)
	Valor() int // a operação "dentro" do elemento (a forma orientada a objetos)
}

type Visitante interface {
	VisitarA(*A)
	VisitarB(*B)
	VisitarC(*C)
}

type A struct{ N int }
type B struct{ N int }
type C struct{ N int }

func (a *A) Aceitar(v Visitante) { v.VisitarA(a) }
func (b *B) Aceitar(v Visitante) { v.VisitarB(b) }
func (c *C) Aceitar(v Visitante) { v.VisitarC(c) }

func (a *A) Valor() int { return a.N }
func (b *B) Valor() int { return b.N * 2 }
func (c *C) Valor() int { return c.N * 3 }

// Soma como visitante clássico.
type Soma struct{ Total int }

func (s *Soma) VisitarA(a *A) { s.Total += a.N }
func (s *Soma) VisitarB(b *B) { s.Total += b.N * 2 }
func (s *Soma) VisitarC(c *C) { s.Total += c.N * 3 }

// Soma como struct de funções.
type SomaFn struct {
	A func(*A)
	B func(*B)
	C func(*C)
}

func (s SomaFn) VisitarA(a *A) { s.A(a) }
func (s SomaFn) VisitarB(b *B) { s.B(b) }
func (s SomaFn) VisitarC(c *C) { s.C(c) }

func somaSwitch(nos []No) int {
	t := 0
	for _, n := range nos {
		switch n := n.(type) {
		case *A:
			t += n.N
		case *B:
			t += n.N * 2
		case *C:
			t += n.N * 3
		}
	}
	return t
}

func main() {
	nos := make([]No, 1000)
	for i := range nos {
		switch i % 3 {
		case 0:
			nos[i] = &A{i}
		case 1:
			nos[i] = &B{i}
		default:
			nos[i] = &C{i}
		}
	}

	var sink int
	correto := &Soma{}
	for _, n := range nos {
		n.Aceitar(correto)
	}
	esperado := correto.Total

	casos := []struct {
		nome string
		f    func() int
	}{
		{"método no elemento (Valor)", func() int {
			t := 0
			for _, n := range nos {
				t += n.Valor()
			}
			return t
		}},
		{"visitante clássico (Aceitar)", func() int {
			s := &Soma{}
			for _, n := range nos {
				n.Aceitar(s)
			}
			return s.Total
		}},
		{"type switch", func() int { return somaSwitch(nos) }},
		{"struct de funções (valor no laço)", func() int {
			t := 0
			s := SomaFn{
				A: func(a *A) { t += a.N },
				B: func(b *B) { t += b.N * 2 },
				C: func(c *C) { t += c.N * 3 },
			}
			for _, n := range nos {
				n.Aceitar(s) // converte SomaFn para Visitante a cada nó
			}
			return t
		}},
		{"struct de funções (ponteiro, 1 conversão)", func() int {
			t := 0
			s := &SomaFn{
				A: func(a *A) { t += a.N },
				B: func(b *B) { t += b.N * 2 },
				C: func(c *C) { t += c.N * 3 },
			}
			for _, n := range nos {
				n.Aceitar(s)
			}
			return t
		}},
	}
	fmt.Println("1000 nós de 3 tipos, somar um valor por tipo:")
	for _, c := range casos {
		if c.f() != esperado {
			fmt.Println("resultado diferente em", c.nome)
		}
		r := testing.Benchmark(func(b *testing.B) {
			for b.Loop() {
				sink = c.f()
			}
		})
		fmt.Printf("  %-30s %7.0f ns/op (%5.2f ns por nó)  %d allocs\n",
			c.nome, float64(r.NsPerOp()), float64(r.NsPerOp())/1000, r.AllocsPerOp())
	}
	_ = sink
}
