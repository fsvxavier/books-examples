package e06

import "testing"

type Forma interface{ Area() float64 }

type Quadrado struct{ Lado float64 }

func (q Quadrado) Area() float64 { return q.Lado * q.Lado }

type Circulo struct{ R float64 }

func (c *Circulo) Area() float64 { return 3 * c.R * c.R }

//go:noinline
func NovoQuadradoConcreto(l float64) Quadrado { return Quadrado{l} }

//go:noinline
func NovoQuadradoInterface(l float64) Forma { return Quadrado{l} }

//go:noinline
func NovoCirculoInterface(r float64) Forma { return &Circulo{r} }

//go:noinline
func Fabrica(tipo int, v float64) Forma {
	switch tipo {
	case 0:
		return Quadrado{v}
	default:
		return &Circulo{v}
	}
}

var sink float64

func BenchmarkConcreto(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = NovoQuadradoConcreto(3).Area()
	}
}

func BenchmarkInterfaceValor(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = NovoQuadradoInterface(3).Area()
	}
}

func BenchmarkInterfacePonteiro(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = NovoCirculoInterface(3).Area()
	}
}

func BenchmarkFabricaComSwitch(b *testing.B) {
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		sink = Fabrica(i&1, 3).Area()
	}
}
