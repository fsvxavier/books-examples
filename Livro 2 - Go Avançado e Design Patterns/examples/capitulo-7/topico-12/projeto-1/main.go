package main

import "fmt"

type Conexao struct{ id int }

func (c *Conexao) Close() error { return nil }

func main() {
	n := 0
	c := NovoCompartilhado(func() (*Conexao, error) { n++; return &Conexao{id: n}, nil })

	a, soltarA, _ := c.Adquirir()
	b, soltarB, _ := c.Adquirir()
	fmt.Printf("dois clientes, mesma conexao: %v | aberturas=%d clientes=%d\n", a == b, c.Aberturas, c.Clientes())

	soltarA()
	soltarA() // duplicado: nao conta duas vezes
	fmt.Printf("A soltou (2x): clientes=%d fechadas=%d\n", c.Clientes(), c.Fechadas)
	soltarB()
	fmt.Printf("B soltou: clientes=%d fechadas=%d\n", c.Clientes(), c.Fechadas)

	d, soltarD, _ := c.Adquirir()
	fmt.Printf("novo cliente reabre: conexao %d | aberturas=%d\n", d.id, c.Aberturas)
	soltarD()
}
