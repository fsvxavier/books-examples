// Package regras avalia regras de preço montadas como árvores de nós.
package regras

import (
	"math"
	"math/rand/v2"
)

// Pedido é a entrada de uma avaliação.
type Pedido struct {
	Itens    int
	Subtotal float64
	Cliente  Nivel
	Cupom    bool
}

// Nivel classifica o cliente.
type Nivel int

// Níveis de cliente.
const (
	Comum Nivel = iota
	Prata
	Ouro
)

// No é um nó da árvore de regras. A chamada Valor, feita por interface,
// é o ponto quente do pacote.
type No interface {
	Valor(p *Pedido) float64
}

// Const devolve sempre o mesmo número.
type Const float64

// Valor implementa No.
func (c Const) Valor(*Pedido) float64 { return float64(c) }

// Subtotal devolve o subtotal do pedido.
type Subtotal struct{}

// Valor implementa No.
func (Subtotal) Valor(p *Pedido) float64 { return p.Subtotal }

// Itens devolve a quantidade de itens.
type Itens struct{}

// Valor implementa No.
func (Itens) Valor(p *Pedido) float64 { return float64(p.Itens) }

// Soma soma dois nós.
type Soma struct{ A, B No }

// Valor implementa No.
func (s Soma) Valor(p *Pedido) float64 { return s.A.Valor(p) + s.B.Valor(p) }

// Mult multiplica dois nós.
type Mult struct{ A, B No }

// Valor implementa No.
func (m Mult) Valor(p *Pedido) float64 { return m.A.Valor(p) * m.B.Valor(p) }

// Minimo escolhe o menor entre dois nós.
type Minimo struct{ A, B No }

// Valor implementa No.
func (m Minimo) Valor(p *Pedido) float64 { return math.Min(m.A.Valor(p), m.B.Valor(p)) }

// Se escolhe um ramo conforme a condição.
type Se struct {
	Cond       func(*Pedido) bool
	Entao, Sen No
}

// Valor implementa No.
func (s Se) Valor(p *Pedido) float64 {
	if s.Cond(p) {
		return s.Entao.Valor(p)
	}
	return s.Sen.Valor(p)
}

// Faixa aplica uma alíquota por faixa de valor. É maior que o orçamento de
// inlining (80), então o compilador não a expande sozinho.
type Faixa struct {
	Base   No
	Cortes [4]float64
	Taxas  [5]float64
}

// Valor implementa No.
func (f Faixa) Valor(p *Pedido) float64 {
	v := f.Base.Valor(p)
	var total, anterior float64
	for i, corte := range f.Cortes {
		if v <= corte {
			total += (v - anterior) * f.Taxas[i]
			return total
		}
		total += (corte - anterior) * f.Taxas[i]
		anterior = corte
	}
	total += (v - anterior) * f.Taxas[len(f.Taxas)-1]
	return total
}

// Preco monta a regra de preço final usada pelo serviço.
func Preco() No {
	nivel := func(n Nivel) func(*Pedido) bool {
		return func(p *Pedido) bool { return p.Cliente == n }
	}
	cupom := func(p *Pedido) bool { return p.Cupom }
	bruto := Subtotal{}
	frete := Minimo{Mult{Itens{}, Const(2.5)}, Const(30)}
	desconto := Se{nivel(Ouro), Mult{bruto, Const(0.10)},
		Se{nivel(Prata), Mult{bruto, Const(0.05)}, Const(0)}}
	extra := Se{cupom, Const(7), Const(0)}
	imposto := Faixa{
		Base:   bruto,
		Cortes: [4]float64{100, 500, 2000, 10000},
		Taxas:  [5]float64{0.02, 0.04, 0.06, 0.08, 0.10},
	}
	return Soma{Soma{bruto, imposto}, Soma{frete, Mult{Soma{desconto, extra}, Const(-1)}}}
}

// Avaliar calcula o preço de um pedido com a regra.
func Avaliar(r No, p *Pedido) float64 { return r.Valor(p) }

// Simular avalia n pedidos sintéticos e devolve a soma dos preços. É o que o
// endpoint /lote do serviço executa para projetar o faturamento de uma campanha.
func Simular(r No, n int, semente uint64) float64 {
	rng := rand.New(rand.NewPCG(semente, 99))
	var total float64
	for range n {
		p := Pedido{
			Itens:    1 + rng.IntN(20),
			Subtotal: rng.Float64() * 3000,
			Cliente:  Nivel(rng.IntN(3)),
			Cupom:    rng.IntN(4) == 0,
		}
		total += Avaliar(r, &p)
	}
	return total
}

// SimularReuso faz o mesmo que Simular, mas reaproveita um único Pedido: o
// ponteiro continua escapando, só que uma vez por chamada, não uma por pedido.
func SimularReuso(r No, n int, semente uint64) float64 {
	rng := rand.New(rand.NewPCG(semente, 99))
	p := new(Pedido)
	var total float64
	for range n {
		*p = Pedido{
			Itens:    1 + rng.IntN(20),
			Subtotal: rng.Float64() * 3000,
			Cliente:  Nivel(rng.IntN(3)),
			Cupom:    rng.IntN(4) == 0,
		}
		total += Avaliar(r, p)
	}
	return total
}
