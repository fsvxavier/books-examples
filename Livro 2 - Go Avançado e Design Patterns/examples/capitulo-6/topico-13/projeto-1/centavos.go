package main

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type Centavos int64

func (c Centavos) String() string {
	sinal := ""
	if c < 0 {
		sinal, c = "-", -c
	}
	return fmt.Sprintf("%sR$ %d,%02d", sinal, c/100, c%100)
}

// Item e o COMPONENTE comum: folha, container E decoradores o implementam.
type Item interface {
	Preco() Centavos
	Descricao() string
}

// ---------- Folha ----------

type Produto struct {
	Nome  string
	Valor Centavos
}

func (p Produto) Preco() Centavos   { return p.Valor }
func (p Produto) Descricao() string { return p.Nome }

// ---------- Container (Composite) ----------

type Kit struct {
	Nome  string
	itens []Item
}

func NovoKit(nome string, itens ...Item) *Kit { return &Kit{nome, itens} }

func (k *Kit) Preco() Centavos {
	var t Centavos
	for _, it := range k.itens {
		t += it.Preco()
	}
	return t
}

func (k *Kit) Descricao() string {
	nomes := make([]string, len(k.itens))
	for i, it := range k.itens {
		nomes[i] = it.Descricao()
	}
	return fmt.Sprintf("%s[%s]", k.Nome, strings.Join(nomes, ", "))
}

// ---------- Decoradores: um UNICO filho, mesma interface ----------

type descontoFixo struct {
	Item
	valor Centavos
}

func DescontoFixo(it Item, valor Centavos) Item { return descontoFixo{it, valor} }

func (d descontoFixo) Preco() Centavos {
	return max(0, d.Item.Preco()-d.valor)
}
func (d descontoFixo) Descricao() string { return d.Item.Descricao() + " -" + d.valor.String() }

type imposto struct {
	Item
	bp int64 // pontos-base: 1000 = 10%
}

func Imposto(it Item, bp int64) Item { return imposto{it, bp} }

func (i imposto) Preco() Centavos {
	base := i.Item.Preco()
	return base + Centavos((int64(base)*i.bp+5000)/10000) // arredonda ao centavo
}
func (i imposto) Descricao() string { return fmt.Sprintf("%s +%d%%", i.Item.Descricao(), i.bp/100) }

// ComCache guarda o preco do item embrulhado. ATENCAO: nao enxerga mudancas feitas depois.
type ComCache struct {
	Item
	calculos atomic.Int64
	pronto   atomic.Bool
	valor    atomic.Int64
}

func (c *ComCache) Preco() Centavos {
	if c.pronto.Load() {
		return Centavos(c.valor.Load())
	}
	v := c.Item.Preco()
	c.calculos.Add(1)
	c.valor.Store(int64(v))
	c.pronto.Store(true)
	return v
}

func (c *ComCache) Invalidar() { c.pronto.Store(false) }

func (k *Kit) Adicionar(it Item) { k.itens = append(k.itens, it) }
