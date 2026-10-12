package main

import (
	"fmt"
	"strings"
)

// Legado: uma impressora com a sua propria interface (nao podemos mudar).
type ImpressoraLegada struct{ Colunas int }

func (p ImpressoraLegada) ImprimirLinha(s string) {
	fmt.Printf("[%d col] %s\n", p.Colunas, s)
}

// ImprimirTudo chama ImprimirLinha DENTRO do legado.
func (p ImpressoraLegada) ImprimirTudo(linhas []string) {
	for _, l := range linhas {
		p.ImprimirLinha(l)
	}
}

// Interface que o nosso codigo espera.
type Saida interface {
	Escrever(texto string) error
}

// "Adapter de classe" aproximado: embute o legado e acrescenta o metodo que falta.
type AdaptadorEmbutido struct {
	ImpressoraLegada
}

func (a AdaptadorEmbutido) Escrever(texto string) error {
	a.ImprimirLinha(strings.TrimSpace(texto))
	return nil
}

// Tentativa de "sobrescrever" um metodo promovido.
func (a AdaptadorEmbutido) ImprimirLinha(s string) {
	fmt.Printf(">>> %s\n", strings.ToUpper(s))
}

// Adapter por composicao: guarda o legado em um campo e expoe so o que quer.
type AdaptadorComposto struct {
	legado ImpressoraLegada
}

func (a AdaptadorComposto) Escrever(texto string) error {
	a.legado.ImprimirLinha(strings.TrimSpace(texto))
	return nil
}

func main() {
	var s1 Saida = AdaptadorEmbutido{ImpressoraLegada{80}}
	var s2 Saida = AdaptadorComposto{ImpressoraLegada{80}}
	s1.Escrever("  ola do embutido  ")
	s2.Escrever("  ola do composto  ")

	fmt.Println("--- ImprimirTudo no embutido: usa a versao sobrescrita?")
	a := AdaptadorEmbutido{ImpressoraLegada{80}}
	a.ImprimirLinha("direto no adaptador")
	a.ImprimirTudo([]string{"via ImprimirTudo"})

	fmt.Println("--- O que o adaptador composto expoe")
	_, ok := s2.(interface{ ImprimirLinha(string) })
	fmt.Println("composto expoe ImprimirLinha:", ok)
	_, ok = s1.(interface{ ImprimirLinha(string) })
	fmt.Println("embutido expoe ImprimirLinha:", ok)
}
