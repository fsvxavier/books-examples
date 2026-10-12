package main

import (
	"fmt"
	"testing"
)

// ---------- State: três estados em ciclo (0 -> 1 -> 2 -> 0) ----------

type Est int

const (
	A Est = iota
	B
	C
)

// 1) switch sobre enum
type MSwitch struct{ e Est }

func (m *MSwitch) Avancar() {
	switch m.e {
	case A:
		m.e = B
	case B:
		m.e = C
	case C:
		m.e = A
	}
}

// 2) interface (um tipo por estado)
type Estado interface{ Avancar(m *MIface) }
type MIface struct{ e Estado }
type eA struct{}
type eB struct{}
type eC struct{}

func (eA) Avancar(m *MIface) { m.e = eB{} }
func (eB) Avancar(m *MIface) { m.e = eC{} }
func (eC) Avancar(m *MIface) { m.e = eA{} }
func (m *MIface) Avancar()   { m.e.Avancar(m) }

// 3) tabela em mapa
type MMapa struct {
	e Est
	t map[Est]Est
}

func (m *MMapa) Avancar() { m.e = m.t[m.e] }

// 4) tabela em array indexado
type MArray struct {
	e Est
	t [3]Est
}

func (m *MArray) Avancar() { m.e = m.t[m.e] }

// 5) função de estado
type MFn struct{ f func(*MFn) }

func fa(m *MFn) { m.f = fb }
func fb(m *MFn) { m.f = fc }
func fc(m *MFn) { m.f = fa }

func (m *MFn) Avancar() { m.f(m) }

// ---------- Strategy: aplicar um desconto ----------

type Tipo int

const (
	Nenhum Tipo = iota
	Percentual
	Fixo
)

func descSwitch(t Tipo, v int) int {
	switch t {
	case Percentual:
		return v - v/10
	case Fixo:
		return v - 500
	}
	return v
}

type Desconto interface{ Aplicar(v int) int }
type perc struct{}
type fixo struct{}

func (perc) Aplicar(v int) int { return v - v/10 }
func (fixo) Aplicar(v int) int { return v - 500 }

type DescontoFunc func(int) int

var sink int

func main() {
	type linha struct {
		nome string
		f    func(b *testing.B)
	}

	fmt.Println("State: uma transição")
	estados := []linha{
		{"switch sobre enum", func(b *testing.B) {
			m := &MSwitch{}
			for b.Loop() {
				m.Avancar()
			}
		}},
		{"interface (1 tipo por estado)", func(b *testing.B) {
			m := &MIface{e: eA{}}
			for b.Loop() {
				m.Avancar()
			}
		}},
		{"tabela em mapa", func(b *testing.B) {
			m := &MMapa{t: map[Est]Est{A: B, B: C, C: A}}
			for b.Loop() {
				m.Avancar()
			}
		}},
		{"tabela em array", func(b *testing.B) {
			m := &MArray{t: [3]Est{B, C, A}}
			for b.Loop() {
				m.Avancar()
			}
		}},
		{"função de estado", func(b *testing.B) {
			m := &MFn{f: fa}
			for b.Loop() {
				m.Avancar()
			}
		}},
	}
	for _, l := range estados {
		r := testing.Benchmark(l.f)
		fmt.Printf("  %-32s %6.2f ns/op  %d allocs\n", l.nome, float64(r.T.Nanoseconds())/float64(r.N), r.AllocsPerOp())
	}

	fmt.Println("Strategy: uma aplicação de desconto")
	estrategias := []linha{
		{"chamada direta", func(b *testing.B) {
			for b.Loop() {
				sink = 10000 - 10000/10
			}
		}},
		{"switch sobre tipo", func(b *testing.B) {
			t := Percentual
			for b.Loop() {
				sink = descSwitch(t, 10000)
			}
		}},
		{"interface", func(b *testing.B) {
			var d Desconto = perc{}
			if len(fmt.Sprint(b.N)) > 100 {
				d = fixo{}
			}
			for b.Loop() {
				sink = d.Aplicar(10000)
			}
		}},
		{"valor função", func(b *testing.B) {
			d := DescontoFunc(func(v int) int { return v - v/10 })
			if len(fmt.Sprint(b.N)) > 100 {
				d = func(v int) int { return v - 500 }
			}
			for b.Loop() {
				sink = d(10000)
			}
		}},
	}
	for _, l := range estrategias {
		r := testing.Benchmark(l.f)
		fmt.Printf("  %-32s %6.2f ns/op  %d allocs\n", l.nome, float64(r.T.Nanoseconds())/float64(r.N), r.AllocsPerOp())
	}
}
