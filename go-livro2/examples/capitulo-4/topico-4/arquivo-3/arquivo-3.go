package main

import (
	"fmt"
	"sync"
)

type Cliente struct {
	once sync.Once
	conn string
}

func (c *Cliente) Conectar() string {
	c.once.Do(func() {
		fmt.Println("  conectando de verdade...")
		c.conn = "conexao-1"
	})
	return c.conn
}

func main() {
	a := &Cliente{}
	a.Conectar() // Once de a consumido

	b := *a // copia o Once junto, ja "usado"
	b.conn = ""
	fmt.Printf("copia b conectou? conn=%q\n", b.Conectar())
}
