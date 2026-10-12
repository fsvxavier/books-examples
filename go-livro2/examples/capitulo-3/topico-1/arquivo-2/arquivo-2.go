package main

import (
	"fmt"
	"strings"
)

// ---- Os produtos: representacoes DIFERENTES, sem interface em comum ----

type Casa struct {
	Janelas, Portas, Andares int
	Material                 string
}

func (c Casa) String() string {
	return fmt.Sprintf("Casa de %s: %d janelas, %d portas, %d andares", c.Material, c.Janelas, c.Portas, c.Andares)
}

type Iglu struct {
	Janelas, Portas, Andares int
	Material                 string
}

func (i Iglu) String() string {
	return fmt.Sprintf("Iglu de %s: %d janela, %d porta", i.Material, i.Janelas, i.Portas)
}

// ---- A interface do builder: os PASSOS de construcao ----

type Construtor interface {
	Paredes()
	Janelas()
	Portas()
	Andares()
}

// ---- Builders concretos: cada um acumula o seu produto ----

type ConstrutorDeCasa struct{ casa Casa }

func NovoConstrutorDeCasa(material string) *ConstrutorDeCasa {
	return &ConstrutorDeCasa{Casa{Material: material}}
}
func (b *ConstrutorDeCasa) Paredes() {}
func (b *ConstrutorDeCasa) Janelas() { b.casa.Janelas = 4 }
func (b *ConstrutorDeCasa) Portas()  { b.casa.Portas = 2 }
func (b *ConstrutorDeCasa) Andares() { b.casa.Andares = 2 }

// Resultado devolve o produto final. O tipo e especifico de cada builder.
func (b *ConstrutorDeCasa) Resultado() Casa { return b.casa }

type ConstrutorDeIglu struct{ iglu Iglu }

func (b *ConstrutorDeIglu) Paredes() { b.iglu.Material = "gelo" }
func (b *ConstrutorDeIglu) Janelas() { b.iglu.Janelas = 1 }
func (b *ConstrutorDeIglu) Portas()  { b.iglu.Portas = 1 }
func (b *ConstrutorDeIglu) Andares() { b.iglu.Andares = 1 }

func (b *ConstrutorDeIglu) Resultado() Iglu { return b.iglu }

// ---- O diretor: sabe a ORDEM e as RECEITAS, nao os detalhes ----

type Diretor struct{}

// Completa constroi tudo.
func (Diretor) Completa(c Construtor) {
	c.Paredes()
	c.Portas()
	c.Janelas()
	c.Andares()
}

// Basica constroi so o minimo.
func (Diretor) Basica(c Construtor) {
	c.Paredes()
	c.Portas()
}

func main() {
	var d Diretor

	pedra := NovoConstrutorDeCasa("pedra")
	d.Completa(pedra)
	fmt.Println(pedra.Resultado())

	madeira := NovoConstrutorDeCasa("madeira")
	d.Basica(madeira)
	fmt.Println(madeira.Resultado())

	iglu := &ConstrutorDeIglu{}
	d.Completa(iglu) // MESMO codigo de construcao, outra representacao
	fmt.Println(iglu.Resultado())

	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("o mesmo Diretor.Completa serviu a dois builders de produtos diferentes")
}
