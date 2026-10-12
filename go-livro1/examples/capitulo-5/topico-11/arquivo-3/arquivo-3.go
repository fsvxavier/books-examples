package main

import (
	"fmt"
	"slices"
)

type Centavos int64

func (c Centavos) String() string { return fmt.Sprintf("R$ %d,%02d", c/100, c%100) }

type Item struct {
	Nome  string
	Preco Centavos
	Qtd   int
}

// Subtotal usa receptor por valor: só lê, e Item é pequeno.
func (i Item) Subtotal() Centavos { return i.Preco * Centavos(i.Qtd) }

type Carrinho struct {
	itens []Item
}

// Adicionar soma a quantidade se o item já existe; senão, acrescenta uma linha nova.
func (c *Carrinho) Adicionar(item Item) {
	for i := range c.itens {
		if c.itens[i].Nome == item.Nome {
			c.itens[i].Qtd += item.Qtd
			return
		}
	}
	c.itens = append(c.itens, item)
}

func (c *Carrinho) Remover(nome string) bool {
	antes := len(c.itens)
	c.itens = slices.DeleteFunc(c.itens, func(i Item) bool { return i.Nome == nome })
	return len(c.itens) < antes
}

// Total usa receptor por ponteiro por consistência com os demais métodos.
func (c *Carrinho) Total() Centavos {
	var total Centavos
	for _, it := range c.itens {
		total += it.Subtotal()
	}
	return total
}

func (c *Carrinho) Itens() []Item { return slices.Clone(c.itens) }

func main() {
	var c Carrinho
	c.Adicionar(Item{"Caderno", 1590, 2})
	c.Adicionar(Item{"Caneta", 350, 5})
	c.Adicionar(Item{"Caderno", 1590, 1})
	fmt.Println(c.Itens(), c.Total())

	fmt.Println(c.Remover("Caneta"), c.Remover("Borracha"))
	fmt.Println(c.Total())

	itens := c.Itens()
	itens[0].Qtd = 99
	fmt.Println(c.Total(), len(c.itens))
}
