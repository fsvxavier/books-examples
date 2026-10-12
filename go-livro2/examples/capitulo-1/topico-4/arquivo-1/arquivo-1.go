package main

import (
	"fmt"
	"sort"
)

// ---- ANTES: cada novo tipo de cliente obriga a EDITAR esta funcao ----

func DescontoSwitch(tipo string, total int64) int64 {
	switch tipo {
	case "regular":
		return 0
	case "ouro":
		return total * 10 / 100
	case "platina":
		return total * 20 / 100
	default:
		return 0
	}
}

// ---- DEPOIS: aberto para extensao (novas regras), fechado para modificacao ----

// Regra e um tipo funcao: o menor "Strategy" possivel em Go.
type Regra func(total int64) int64

// Desconto e um tipo funcao que implementa a interface por conta propria,
// como http.HandlerFunc faz com http.Handler.
type Descontador interface{ Desconto(total int64) int64 }

func (r Regra) Desconto(total int64) int64 { return r(total) }

type Percentual int

func (p Percentual) Desconto(total int64) int64 { return total * int64(p) / 100 }

// Politica guarda as regras por nome. Quem usa nao muda quando uma regra nasce.
type Politica struct{ regras map[string]Descontador }

func NovaPolitica() *Politica { return &Politica{regras: map[string]Descontador{}} }

func (p *Politica) Registrar(nome string, d Descontador) { p.regras[nome] = d }

func (p *Politica) Aplicar(nome string, total int64) int64 {
	if d, ok := p.regras[nome]; ok {
		return total - d.Desconto(total)
	}
	return total
}

func (p *Politica) Nomes() []string {
	n := make([]string, 0, len(p.regras))
	for k := range p.regras {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}

func main() {
	fmt.Println("switch ouro 10000:", DescontoSwitch("ouro", 10000))

	pol := NovaPolitica()
	pol.Registrar("regular", Percentual(0))
	pol.Registrar("ouro", Percentual(10))
	pol.Registrar("platina", Percentual(20))
	fmt.Println("politica:", pol.Nomes())
	fmt.Println("ouro 10000 ->", pol.Aplicar("ouro", 10000))

	// requisito novo: "aniversario" = 15%, com teto de R$ 50,00 de desconto.
	// Nenhuma linha de Politica ou de DescontoSwitch foi editada.
	pol.Registrar("aniversario", Regra(func(total int64) int64 {
		return min(total*15/100, 5000)
	}))
	fmt.Println("politica:", pol.Nomes())
	fmt.Println("aniversario 10000 ->", pol.Aplicar("aniversario", 10000), "(15% = 1500)")
	fmt.Println("aniversario 100000 ->", pol.Aplicar("aniversario", 100000), "(teto de 5000 de desconto)")
	fmt.Println("tipo desconhecido ->", pol.Aplicar("novo", 10000), "(sem desconto)")
}
