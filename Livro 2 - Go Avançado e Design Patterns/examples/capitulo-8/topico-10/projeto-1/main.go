package main

import (
	"encoding/json"
	"fmt"
)

type contador struct {
	pinturas int
	porTipo  map[string]int
}

func (c *contador) Pintar(nome, cor string, tex int, x, y float64) {
	c.pinturas++
	c.porTipo[nome]++
}

func main() {
	catalogo := map[string]TipoDef{
		"carvalho": {"verde-escuro", make([]byte, 2048)},
		"pinheiro": {"verde", make([]byte, 2048)},
		"betula":   {"branco", make([]byte, 2048)},
	}
	nomes := []string{"carvalho", "pinheiro", "betula"}

	f := NovaFabricaTipos()
	fl := NovaFloresta(f)
	for i := range 30000 {
		n := nomes[i%3]
		fl.Plantar(float64(i%200), float64(i/200), n, catalogo[n].Cor, catalogo[n].Textura)
	}
	fmt.Printf("arvores plantadas: %d | flyweights criados: %d\n", fl.Tamanho(), f.Quantidade())

	tela := &contador{porTipo: map[string]int{}}
	fl.Desenhar(tela)
	fmt.Printf("pinturas: %d | por tipo: %v\n", tela.pinturas, tela.porTipo)

	// Persistencia: so contexto + chave
	dados, _ := json.Marshal(fl)
	fmt.Printf("JSON com chave do tipo: %d bytes\n", len(dados))

	// O que aconteceria se serializassemos o flyweight dentro de cada arvore:
	type tipoAberto struct {
		Nome, Cor string
		Textura   []byte
	}
	type arvoreAberta struct {
		X, Y float64
		Tipo tipoAberto
	}
	aberto := make([]arvoreAberta, 0, 3000)
	for i := range 3000 { // so 3000, para nao gastar memoria
		n := nomes[i%3]
		aberto = append(aberto, arvoreAberta{float64(i % 200), float64(i / 200), tipoAberto{n, catalogo[n].Cor, catalogo[n].Textura}})
	}
	bruto, _ := json.Marshal(aberto)
	fmt.Printf("JSON com o tipo repetido em cada arvore (so 3.000 arvores): %d bytes\n", len(bruto))

	// Recarregar: o compartilhamento volta, pela fabrica.
	f2 := NovaFabricaTipos()
	fl2, err := Carregar(dados, f2, catalogo)
	fmt.Printf("recarregada: %d arvores, %d flyweights, erro=%v\n", fl2.Tamanho(), f2.Quantidade(), err)

	_, err = Carregar([]byte(`[{"x":1,"y":2,"tipo":"jacaranda"}]`), NovaFabricaTipos(), catalogo)
	fmt.Println("tipo desconhecido:", err)
}
