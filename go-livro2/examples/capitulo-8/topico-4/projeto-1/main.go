package main

import (
	"errors"
	"fmt"
)

func main() {
	f := NovaFabrica(criarTipo)

	a, _ := f.Obter(Chave{"carvalho", "verde"})
	b, _ := f.Obter(Chave{"carvalho", "verde"})
	c, _ := f.Obter(Chave{"pinheiro", "verde"})
	fmt.Println("mesma chave, mesmo ponteiro:", a == b)
	fmt.Println("chaves diferentes, ponteiros diferentes:", a != c)
	fmt.Println("flyweights criados:", f.Criados, "| no pool:", f.Tamanho())
	fmt.Println(a.Desenhar(10, 20))

	// Alterar a copia da textura NAO afeta o flyweight.
	copia := a.Textura()
	copia[0] = 99
	fmt.Println("flyweight intacto depois de mexer na copia:", a.Textura()[0] == 0)

	_, err := f.Obter(Chave{"", "verde"})
	fmt.Println("chave invalida:", err, "| errors.Is:", errors.Is(err, ErrChave), "| no pool:", f.Tamanho())
}
