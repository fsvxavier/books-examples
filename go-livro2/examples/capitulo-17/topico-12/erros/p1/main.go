package main

import "fmt"

type Iniciavel interface{ Iniciar(string) }

type Conexao struct{ dsn string }

func (c *Conexao) Iniciar(dsn string) { c.dsn = dsn }

func Novo[T Iniciavel](arg string) T {
	var v T
	v.Iniciar(arg)
	return v
}

func main() { fmt.Println(Novo[Conexao]("x")) }
