package main

import (
	"fmt"
	"strings"
)

// Centavos evita ponto flutuante em dinheiro.
type Centavos int64

func (c Centavos) String() string { return fmt.Sprintf("R$ %d,%02d", c/100, c%100) }

// Componente: o que serve tanto para uma folha quanto para um container.
type Item interface {
	Preco() Centavos
	Descrever(nivel int) string
}

// Folha: nao tem filhos.
type Produto struct {
	Nome  string
	Valor Centavos
}

func (p Produto) Preco() Centavos { return p.Valor }
func (p Produto) Descrever(nivel int) string {
	return fmt.Sprintf("%s- %s (%v)\n", strings.Repeat("  ", nivel), p.Nome, p.Valor)
}

// Container: guarda Itens, que podem ser folhas OU outros containers.
type Caixa struct {
	Nome      string
	Embalagem Centavos
	itens     []Item
}

func NovaCaixa(nome string, embalagem Centavos, itens ...Item) *Caixa {
	return &Caixa{Nome: nome, Embalagem: embalagem, itens: itens}
}

func (c *Caixa) Adicionar(it Item) { c.itens = append(c.itens, it) }

func (c *Caixa) Preco() Centavos {
	total := c.Embalagem
	for _, it := range c.itens {
		total += it.Preco() // o container NAO sabe se o filho e folha ou caixa
	}
	return total
}

func (c *Caixa) Descrever(nivel int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s+ caixa %s (embalagem %v, total %v)\n", strings.Repeat("  ", nivel), c.Nome, c.Embalagem, c.Preco())
	for _, it := range c.itens {
		b.WriteString(it.Descrever(nivel + 1))
	}
	return b.String()
}

func main() {
	pedido := NovaCaixa("pedido", 500,
		Produto{"Fone de ouvido", 19990},
		NovaCaixa("kit escritorio", 300,
			Produto{"Caneta", 450},
			Produto{"Caderno", 2990},
			NovaCaixa("brindes", 0,
				Produto{"Adesivo", 100},
				Produto{"Chaveiro", 350},
			),
		),
		Produto{"Mouse", 8900},
	)

	fmt.Print(pedido.Descrever(0))
	fmt.Println("total do pedido:", pedido.Preco())

	// O cliente trata folha e caixa do mesmo jeito.
	var qualquer Item = Produto{"Cabo", 1500}
	fmt.Println("preco de um item solto:", qualquer.Preco())
	qualquer = pedido
	fmt.Println("preco do pedido inteiro:", qualquer.Preco())
}
