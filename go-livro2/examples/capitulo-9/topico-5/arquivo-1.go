package main

import (
	"fmt"
	"strings"
)

type Evento struct {
	Tipo string // "clique", "tecla:ESC", "tecla:CTRL+S"
}

// Componente faz parte de uma ARVORE (Composite): cada um conhece o seu pai.
type Componente struct {
	Nome  string
	Pai   *Componente
	Trata map[string]func(Evento) // eventos que ESTE componente sabe tratar
}

func Novo(nome string, pai *Componente, trata map[string]func(Evento)) *Componente {
	return &Componente{Nome: nome, Pai: pai, Trata: trata}
}

// Disparar entrega o evento ao componente e, se ele nao o trata, SOBE pela cadeia de pais
// ate a raiz. O primeiro que sabe tratar, trata e encerra. Devolve quem tratou ("" se ninguem).
func (c *Componente) Disparar(e Evento, caminho *[]string) string {
	for atual := c; atual != nil; atual = atual.Pai {
		*caminho = append(*caminho, atual.Nome)
		if f, ok := atual.Trata[e.Tipo]; ok {
			f(e)
			return atual.Nome
		}
	}
	return ""
}

func main() {
	var log []string
	janela := Novo("janela", nil, map[string]func(Evento){
		"tecla:CTRL+S": func(Evento) { log = append(log, "salvou o documento") },
		"tecla:ESC":    func(Evento) { log = append(log, "fechou a janela") },
	})
	painel := Novo("painel", janela, map[string]func(Evento){
		"tecla:ESC": func(Evento) { log = append(log, "painel fechou o menu aberto") },
	})
	botao := Novo("botao", painel, map[string]func(Evento){
		"clique": func(Evento) { log = append(log, "botao: acao do clique") },
	})

	for _, e := range []Evento{{"clique"}, {"tecla:ESC"}, {"tecla:CTRL+S"}, {"tecla:F9"}} {
		var caminho []string
		quem := botao.Disparar(e, &caminho)
		fmt.Printf("%-13s caminho: %-24s tratado por: %q\n", e.Tipo, strings.Join(caminho, " > "), quem)
	}
	for _, l := range log {
		fmt.Println("  efeito:", l)
	}
}
