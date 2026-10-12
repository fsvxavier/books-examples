package main

import (
	"fmt"
	"maps"
	"slices"
)

type Endereco struct{ Rua, Cidade string }

type Cliente struct {
	Nome     string
	Tags     []string          // slice: cabecalho aponta para um array compartilhado
	Metas    map[string]string // mapa: referencia compartilhada
	Endereco *Endereco         // ponteiro: aponta para o MESMO endereco
	Casa     Endereco          // valor: copiado de verdade
}

// ---- 1. a copia "ingenua": atribuicao de struct = copia RASA ----

func demoRasa() {
	a := Cliente{
		Nome: "Ana", Tags: []string{"vip"}, Metas: map[string]string{"plano": "ouro"},
		Endereco: &Endereco{"Rua A", "Recife"}, Casa: Endereco{"Rua B", "Olinda"},
	}
	b := a // copia rasa

	b.Nome = "Bia"                 // independente (string e valor)
	b.Casa.Rua = "Rua Z"           // independente (struct por valor)
	b.Tags[0] = "comum"            // COMPARTILHADO (mesmo array)
	b.Metas["plano"] = "bronze"    // COMPARTILHADO (mesmo mapa)
	b.Endereco.Cidade = "Salvador" // COMPARTILHADO (mesmo ponteiro)

	fmt.Println("1. copia rasa (b := a), depois de alterar b:")
	fmt.Printf("   a.Nome=%s (independente)\n", a.Nome)
	fmt.Printf("   a.Casa.Rua=%s (independente)\n", a.Casa.Rua)
	fmt.Printf("   a.Tags=%v (COMPARTILHADO)\n", a.Tags)
	fmt.Printf("   a.Metas=%v (COMPARTILHADO)\n", a.Metas)
	fmt.Printf("   a.Endereco.Cidade=%s (COMPARTILHADO)\n", a.Endereco.Cidade)
}

// ---- 2. Clone profundo, escrito a mao ----

func (c Cliente) Clone() Cliente {
	novo := c // copia os campos de valor
	novo.Tags = slices.Clone(c.Tags)
	novo.Metas = maps.Clone(c.Metas)
	if c.Endereco != nil {
		e := *c.Endereco
		novo.Endereco = &e
	}
	return novo
}

func demoProfunda() {
	a := Cliente{
		Nome: "Ana", Tags: []string{"vip"}, Metas: map[string]string{"plano": "ouro"},
		Endereco: &Endereco{"Rua A", "Recife"},
	}
	b := a.Clone()
	b.Tags[0] = "comum"
	b.Metas["plano"] = "bronze"
	b.Endereco.Cidade = "Salvador"
	fmt.Println("2. Clone() profundo, depois de alterar b:")
	fmt.Printf("   a.Tags=%v a.Metas=%v a.Endereco.Cidade=%s  (nada mudou)\n", a.Tags, a.Metas, a.Endereco.Cidade)
}

// ---- 3. slices.Clone e maps.Clone sao RASOS nos elementos ----

func demoRasoNosElementos() {
	m := map[string][]int{"a": {1, 2, 3}}
	c := maps.Clone(m)
	c["a"][0] = 99 // o slice DENTRO do mapa e o mesmo
	c["b"] = []int{7}
	fmt.Println("3. maps.Clone de map[string][]int:")
	fmt.Printf("   original m=%v (o slice interno foi alterado; a chave 'b' nao apareceu)\n", m)

	ptrs := []*Endereco{{"R1", "X"}}
	cp := slices.Clone(ptrs)
	cp[0].Cidade = "Y"
	fmt.Printf("   slices.Clone de []*Endereco: original=%v (mesmo ponteiro)\n", *ptrs[0])
}

// ---- 4. slices: o append depois da copia pode ou nao afetar o original ----

func demoAppend() {
	a := make([]int, 3, 10)
	b := a[:3]        // mesmo array, capacidade sobrando
	b = append(b, 42) // escreve no array compartilhado
	c := append(a, 7) // sobrescreve o MESMO slot!
	fmt.Printf("4. append em slices que compartilham array: b[3]=%d c[3]=%d (b foi sobrescrito)\n", b[3], c[3])
	d := slices.Clone(a)
	d = append(d, 1)
	fmt.Printf("   com slices.Clone o array e novo: a[:4][3]=%d\n", a[:4][3])
}

func main() {
	demoRasa()
	demoProfunda()
	demoRasoNosElementos()
	demoAppend()
}
