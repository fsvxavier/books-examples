package main

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sort"
)

// Ponto é uma coordenada em quilômetros.
type Ponto struct{ X, Y float64 }

func (a Ponto) Dist(b Ponto) float64 { return math.Hypot(a.X-b.X, a.Y-b.Y) }

// Rota é o resultado de um cálculo: custo em centavos, tempo em minutos.
type Rota struct {
	Modo    string
	Km      float64
	Minutos float64
	Custo   int
}

// Estrategia é a interface do catálogo: um algoritmo de roteamento.
type Estrategia interface {
	Calcular(de, para Ponto) Rota
}

// EstrategiaFunc deixa uma função ser uma Estrategia.
type EstrategiaFunc func(de, para Ponto) Rota

func (f EstrategiaFunc) Calcular(de, para Ponto) Rota { return f(de, para) }

// Forma 1: um tipo por algoritmo, com configuração (campos).
type PorVelocidade struct {
	Modo        string
	Sinuosidade float64 // distância real = reta x sinuosidade
	KmPorHora   float64
	EsperaMin   float64
	TarifaFixa  int
	CentavosKm  int
}

func (p PorVelocidade) Calcular(de, para Ponto) Rota {
	km := de.Dist(para) * p.Sinuosidade
	return Rota{
		Modo:    p.Modo,
		Km:      km,
		Minutos: km/p.KmPorHora*60 + p.EsperaMin,
		Custo:   p.TarifaFixa + int(math.Round(km*float64(p.CentavosKm))),
	}
}

// Forma 2: funções comuns (sem tipo novo).
func aPe(de, para Ponto) Rota {
	km := de.Dist(para) * 1.2
	return Rota{"a pé", km, km / 5 * 60, 0}
}

// Forma 3: tabela de registro, para escolher por nome (vindo da tela, de
// uma configuração, de um parâmetro de URL).
var estrategias = map[string]Estrategia{
	"carro":      PorVelocidade{"carro", 1.3, 40, 0, 0, 90},
	"bicicleta":  PorVelocidade{"bicicleta", 1.2, 15, 0, 0, 0},
	"transporte": PorVelocidade{"transporte", 1.4, 25, 8, 450, 0},
	"a pé":       EstrategiaFunc(aPe),
}

// Navegador é o contexto: guarda a estratégia e delega.
type Navegador struct {
	estrategia Estrategia
}

func (n *Navegador) Definir(e Estrategia) { n.estrategia = e }

func (n *Navegador) Rota(de, para Ponto) Rota { return n.estrategia.Calcular(de, para) }

// Melhor escolhe entre TODAS as estratégias com um critério, que por sua
// vez também é uma estratégia (uma função de comparação).
func Melhor(de, para Ponto, criterio func(a, b Rota) int) Rota {
	var rotas []Rota
	for _, e := range estrategias {
		rotas = append(rotas, e.Calcular(de, para))
	}
	return slices.MinFunc(rotas, criterio)
}

func (r Rota) String() string {
	return fmt.Sprintf("%-10s %5.1f km %6.1f min R$ %5.2f", r.Modo, r.Km, r.Minutos, float64(r.Custo)/100)
}

func main() {
	casa, trabalho := Ponto{0, 0}, Ponto{6, 8} // 10 km em linha reta

	fmt.Println("1) o cliente escolhe a estratégia pelo nome")
	nav := &Navegador{}
	nomes := make([]string, 0, len(estrategias))
	for n := range estrategias {
		nomes = append(nomes, n)
	}
	sort.Strings(nomes)
	for _, nome := range nomes {
		nav.Definir(estrategias[nome])
		fmt.Println("  ", nav.Rota(casa, trabalho))
	}

	fmt.Println("2) uma estratégia ad hoc, só com uma função")
	nav.Definir(EstrategiaFunc(func(de, para Ponto) Rota {
		return Rota{"helicóptero", de.Dist(para), de.Dist(para) / 200 * 60, 120000}
	}))
	fmt.Println("  ", nav.Rota(casa, trabalho))

	fmt.Println("3) o critério também é uma estratégia")
	maisRapido := func(a, b Rota) int { return cmp.Compare(a.Minutos, b.Minutos) }
	maisBarato := func(a, b Rota) int {
		return cmp.Or(cmp.Compare(a.Custo, b.Custo), cmp.Compare(a.Minutos, b.Minutos))
	}
	fmt.Println("   mais rápido:", Melhor(casa, trabalho, maisRapido))
	fmt.Println("   mais barato:", Melhor(casa, trabalho, maisBarato))
}
