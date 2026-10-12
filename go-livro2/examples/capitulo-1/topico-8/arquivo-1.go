package main

import (
	"fmt"
)

// ---- Superengenharia: 4 tipos e 2 interfaces para dar bom-dia ----

type Saudador interface{ Saudar(nome string) string }

type FabricaDeSaudador interface{ Criar() Saudador }

type saudadorPadrao struct{ prefixo string }

func (s saudadorPadrao) Saudar(nome string) string { return s.prefixo + ", " + nome }

type fabricaPadrao struct{}

func (fabricaPadrao) Criar() Saudador { return saudadorPadrao{prefixo: "Bom dia"} }

type ServicoDeSaudacao struct{ fabrica FabricaDeSaudador }

func NovoServicoDeSaudacao(f FabricaDeSaudador) *ServicoDeSaudacao { return &ServicoDeSaudacao{f} }

func (s *ServicoDeSaudacao) Executar(nome string) string { return s.fabrica.Criar().Saudar(nome) }

// ---- A solucao direta ----

func Saudar(nome string) string { return "Bom dia, " + nome }

func main() {
	complicado := NovoServicoDeSaudacao(fabricaPadrao{}).Executar("Ana")
	direto := Saudar("Ana")
	fmt.Println(complicado)
	fmt.Println(direto)
	fmt.Println("mesmo resultado:", complicado == direto)
}
